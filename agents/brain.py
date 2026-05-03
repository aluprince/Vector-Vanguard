import os
import asyncio
import logging
import sys
from dotenv import load_dotenv, find_dotenv
from aiogram import Bot, Dispatcher, html
from aiogram.client.default import DefaultBotProperties
from aiogram.enums import ParseMode
from aiogram.filters import CommandStart
from aiogram.types import CallbackQuery, Message
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton

from agents.prompts import generate_pitches, intro_verification_greetings
from aiohttp import web

load_dotenv(find_dotenv())


# Bot token can be obtained via https://t.me/BotFather
TOKEN = os.getenv("BOT_TOKEN")
bot = Bot(token=TOKEN, default=DefaultBotProperties(parse_mode=ParseMode.HTML))
leads = [] # This will hold the leads data received from Go
leads_length = len(leads)

# All handlers should be attached to the Router (or Dispatcher)
dp = Dispatcher()
users = set() 

@dp.message(CommandStart())
async def command_start_handler(message: Message) -> None:
    """
    This handler receives messages with `/start` command
    """
    # Most event objects have aliases for API methods that can be called in events' context
    # For example if you want to answer to incoming message you can use `message.answer(...)` alias
    # and the target chat will be passed to :ref:`aiogram.methods.send_message.SendMessage`
    # method automatically or call API method directly via
    # Bot instance: `bot.send_message(chat_id=message.chat.id, ...)`
    users.add(message.chat.id)
    await message.answer(f"Hello, {html.bold(message.from_user.full_name)}!, chat_id: {message.chat.id}")


@dp.message()
async def talk_to_bot(message: Message) -> None:
    if message.text == "Hello":
        await message.answer("Hi there! How can I assist you today?")
    elif message.text == "What's your name?":
        await message.answer("I'm Vector Vanguard, your technical consultant bot!")

@dp.message()
async def get_chat_id(message: Message):
    if message.text.lower() == "id":
        print(message.chat.id)
        await message.answer(f"Your chat_id is: {message.chat.id}")

@dp.callback_query(lambda c: c.data.startswith("pitch_"))
async def send_pitch_callback(callback_query: CallbackQuery):
    if callback_query.data.startswith("pitch_"):
        lead_id = int(callback_query.data.split("_")[1])
        lead = next((lead for lead in leads if lead['id'] == lead_id), None)
        if lead:
            await callback_query.message.answer(f"📢 *Pitch for {lead['business_name']}*\n\n{lead['pitch']}", parse_mode="Markdown")
            await callback_query.answer()  # Acknowledge the callback query
        else:
            await callback_query.message.answer("Sorry, I couldn't find the pitch for this lead.")
            await callback_query.answer()

@dp.callback_query(lambda c: c.data.startswith("greet_"))
async def send_greeting_callback(callback_query: CallbackQuery):
    if callback_query.data.startswith("greet_"):
        lead_id = int(callback_query.data.split("_")[1])
        lead = next((lead for lead in leads if lead['id'] == lead_id), None)
        if lead:
            lead_greeting = intro_verification_greetings(lead['business_name'])
            await callback_query.message.answer(f"👋 *Greeting for {lead['business_name']}*\n\n{lead_greeting}", parse_mode="Markdown")
            await callback_query.answer()
        else:
            await callback_query.message.answer("Sorry, I couldn't find the greeting for this lead.")
            await callback_query.answer()

async def broadcast_leads():
    pitches_sent = 0
    while pitches_sent < leads_length:
        for lead in leads:
            print(f">> NEW LEAD: {lead}")
            text = (
                f"🚀 *New Lead Alert!*\n"
                f"Lead ID: {lead['id']}\n"
                f"🏨 *Lead:* {lead['business_name']}\n"
                f"📱 *Phone:* {lead['phone']}\n"
                f"💡 *Pitch:* {lead['pitch']}\n"
                # f"❌ *Issue:* {lead['failure_reason']}\n"
                f"⚖️ *Audit Score:* {lead['audit_score']}"
            )
            phone = lead['phone']
            wa_phone = phone.lstrip('0')

            keyboard = InlineKeyboardMarkup(inline_keyboard=[
                [InlineKeyboardButton(text="📱 Message via WhatsApp", url=f"https://wa.me/{wa_phone}")],
                [InlineKeyboardButton(text="📝 Get Pitch", callback_data=f"pitch_{lead['id']}")],
                [InlineKeyboardButton(text="👋 Greeting", callback_data=f"greet_{lead['id']}")]
            ])
            try:
                # print(f"USER ID: {user_id}")
                await bot.send_message("6722182179", text=text, reply_markup=keyboard, parse_mode="Markdown")
                await asyncio.sleep(30)
                pitches_sent += 1
            except Exception as e:
                print(f"error: {e}")
                pass
        await asyncio.sleep(10)

async def handle_webhook(request):
    """This receives the JSON from Go"""
    global leads, leads_length
    data = await request.json()
    print(f"📥 Received {len(data)} leads from Go!")
    
    # Run the pitch generation logic on the new data
    # Note: Modify prompts.py to accept data as an argument instead of reading file
    leads = generate_pitches(data) 
    leads_length = len(leads)
    
    # Trigger the broadcast immediately
    asyncio.create_task(broadcast_leads())
    
    return web.Response(text="Leads received and processing started")


async def main() -> None:
    # Setup the Webhook Server
    app = web.Application()
    app.router.add_post('/new_leads', handle_webhook)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, 'localhost', 5000)
    
    # Start the Web Server and the Bot Polling together
    await asyncio.gather(
        site.start(),
        dp.start_polling(bot)
    )

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, stream=sys.stdout)
    asyncio.run(main())