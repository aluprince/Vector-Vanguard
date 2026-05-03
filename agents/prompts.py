import re
import os
import json
from langchain_groq import ChatGroq
from langchain_core.prompts import PromptTemplate
from dotenv import load_dotenv, find_dotenv
import random

load_dotenv(find_dotenv())

groq_api_key = os.getenv("GROQ_API_KEY_44")
model = os.getenv("AI_MODEL_2")

# passing the audit data and business name into the prompt template
template = """
You are a sharp, professional technical consultant in Lagos. You are messaging a business owner on WhatsApp.
Write in simple, clear English. No tech jargon. No big words like "establish" or "verify."

Business: {business_name}
Audit Problem: {audit_json}

STRICT STYLE:
- Start with "Good day."
- Speak like a human being. Use "I noticed" or "I was checking."
- Explain the "Salt": If people don't see a website, they don't trust the business. They will go and book somewhere else.
- Use words like "Proper," "Standard," "Bookings," and "Money."
- If the business uses facebook, whatsapp, or instagram as their main online presence, call it out as "Not Proper" and "Not Standard."
- Keep it under 70 words.


EXAMPLES OF THE TONE:
"Good day. I was trying to check out your shortlet online but noticed you don't have a website. Honestly, many people won't book if they can't find a site to confirm you are standard. You're likely losing a lot of clients to others. I'm a developer and can help you fix this quickly."

"Good day. I noticed your website is currently down. Directly speaking, this is costing you money because clients who see a broken link will just move to the next apartment. It makes the business look unattended. I can help you get it back online today."

YOUR TURN:
"""


# Initialize Groq
llm = ChatGroq(temperature=0.7, model_name=model, groq_api_key=groq_api_key)

# Opening the Json Schema file and loading it into a variable
def generate_pitches(audit_data):
    # with open("final_audit.json", "r") as file:
    #     audit_data = json.load(file)
    
        leads_dict = []

        for i in range(len(audit_data)):
            print(f"Processing audit {i+1}/{len(audit_data)}...")
            audit_json = audit_data[i]
            business_name = audit_json.get("name", "Unknown Business")  # Extracting business name from the audit data
            if business_name == "":
                business_name = "Unknown Business"

            prompt = PromptTemplate.from_template(template)
            chain = prompt | llm
            pitch = chain.invoke({
                "audit_json": json.dumps(audit_json),
                "business_name": business_name
            })

            raw_phone = audit_json.get("phone", "")
            clean_phone = "".join(re.findall(r'\d+', raw_phone))

            leads_dict.append({
                "id": i+1,
                "business_name": business_name,
                "phone": clean_phone,
                "audit": audit_json,
                "pitch": pitch.content,
                "audit_score": audit_json.get("score", 0)
            })
            # print(f"Generated Pitch {i+1}/{len(audit_data)}: {pitch.content}")
            print(leads_dict)
        return leads_dict


def intro_verification_greetings(name):
    intros = [
    f"Hi, I'm I speaking with the manager of {name}?",
    f"Hello! Is this the official line for {name} shortlets?",
    f"Hey, quick one—do you guys handle the bookings for {name}?",
    f"Good day, wanted to confirm if this is {name} in Lagos?"
]
    if name == "Unknown Business" or name == "" or name is None:
        return "Hi, I'm I speaking with the manager?"
    selected_intro = random.choice(intros)

    return selected_intro