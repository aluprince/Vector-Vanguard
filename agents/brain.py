import os
import json
import asyncio
import logging
import sys
import random
from dotenv import load_dotenv, find_dotenv
from aiogram import Bot, Dispatcher, html
from aiogram.client.default import DefaultBotProperties
from aiogram.enums import ParseMode
from aiogram.filters import CommandStart
from aiogram.types import CallbackQuery, Message
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from agents.prompts import generate_pitches, intro_verification_greetings
from aiohttp import web
from agents.database import init_db, save_lead, get_next_pending_lead, mark_as_sent, get_lead_by_id

load_dotenv(find_dotenv())


# Bot token can be obtained via https://t.me/BotFather
broadcaster_running = False
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
    lead_id = int(callback_query.data.split("_")[1])
    
    # Check DB instead of the volatile 'leads' list
    lead = get_lead_by_id(lead_id) 
    
    if lead:
        await callback_query.message.answer(
            f"📢 *Pitch for {lead['business_name']}*\n\n{lead['pitch']}", 
            parse_mode="Markdown"
        )
    else:
        await callback_query.message.answer("Lead not found in database.")
    await callback_query.answer()

@dp.callback_query(lambda c: c.data.startswith("greet_"))
async def send_greeting_callback(callback_query: CallbackQuery):
    if callback_query.data.startswith("greet_"):
        lead_id = int(callback_query.data.split("_")[1])
        lead = get_lead_by_id(lead_id)
        if lead:
            lead_greeting = intro_verification_greetings(lead['business_name'])
            await callback_query.message.answer(f"👋 *Greeting for {lead['business_name']}*\n\n{lead_greeting}", parse_mode="Markdown")
            await callback_query.answer()
        else:
            await callback_query.message.answer("Sorry, I couldn't find the greeting for this lead.")
            await callback_query.answer()

async def broadcast_leads():
    print("📡 Broadcaster started: Checking for pending leads...")
    
    while True:
        lead = get_next_pending_lead()
        
        if not lead:
            print("🏁 All pending leads have been sent. Broadcaster going to sleep.")
            break  # Exit the loop if nothing is left to send
            
        print(f">> SENDING LEAD: {lead['business_name']}")
        
        # Note: lead is an sqlite3.Row object, access it like a dictionary
        text = (
            f"🚀 *New Lead Alert!*\n"
            f"Lead ID: {lead['id']}\n"
            f"🏨 *Lead:* {lead['business_name']}\n"
            f"📱 *Phone:* {lead['phone']}\n"
            f"💡 *Pitch:* {lead['pitch']}\n"
            f"⚖️ *Audit Score:* {json.loads(lead['audit_json']).get('score', 0)}"
        )
        
        wa_phone = lead['phone'].lstrip('0')

        keyboard = InlineKeyboardMarkup(inline_keyboard=[
            [InlineKeyboardButton(text="📱 Message via WhatsApp", url=f"https://wa.me/{wa_phone}")],
            [InlineKeyboardButton(text="📝 Get Pitch", callback_data=f"pitch_{lead['id']}")],
            [InlineKeyboardButton(text="👋 Greeting", callback_data=f"greet_{lead['id']}")]
        ])

        try:
            await bot.send_message("6722182179", text=text, reply_markup=keyboard, parse_mode="Markdown")
            
            mark_as_sent(lead['id'])
            
            sleep_time = 10 # For Testing
            # sleep_time = random.randint(1200, 1800) # For 20-30 minutes sleep
            print(f"✅ Message sent. Sleeping for {sleep_time // 60} minutes...")
            await asyncio.sleep(sleep_time) 
            
        except Exception as e:
            print(f"⚠️ Error sending message: {e}")
            await asyncio.sleep(60) # Short sleep before retrying same lead


async def handle_webhook(request):
    """This receives the JSON from Go"""
    global leads, leads_length, broadcaster_running
    data = await request.json()
    print(f"📥 Received {len(data)} leads from Go!")
    
    # Run the pitch generation logic on the new data
    leads = generate_pitches(data) 
    leads_length = len(leads)
    for lead in leads:
        save_lead(lead) # Save each lead to the database
    # Trigger the broadcast immediately
    if not broadcaster_running:
        broadcaster_running = True
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
    init_db()
    logging.basicConfig(level=logging.INFO, stream=sys.stdout)
    asyncio.run(main())