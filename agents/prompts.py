from langchain_groq import ChatGroq
from langchain.prompts import PromptTemplate
from dotenv import load_dotenv, find_dotenv
import random

load_dotenv(find_dotenv())

# Initialize Groq
llm = ChatGroq(temperature=0.7, model_name="llama3-70b-8192", groq_api_key="your_key")

template = """
You are a direct, tech-savvy local freelancer in Nigeria. 
Audit Data: {audit_json}
Business: {business_name}

Task: Write a short, casual WhatsApp pitch. 
1. Mention a specific technical pain point from the audit.
2. "Salt the wound" by explaining how it's losing them money.
3. Suggest a fix without being "salesy".
4. NO corporate greetings. Use a "hey" or just start with the observation.
"""

# This chain generates your "Entropy" pitch
prompt = PromptTemplate.from_template(template)
chain = prompt | llm



def intro_verification_greetings(name):
    intros = [
    f"Hi, am I speaking with the manager of {name}?",
    f"Hello! Is this the official line for {name} shortlets?",
    f"Hey, quick one—do you guys handle the bookings for {name}?",
    f"Good day, wanted to confirm if this is {name} in Lagos?"
]
    selected_intro = random.choice(intros)

    return selected_intro



# system_prompt = """
# You are a local tech freelancer in Lagos. 
# STRICT RULES:
# - NO 'Dear', 'I hope this finds you well', or 'Best regards'.
# - NO 'Revolutionize', 'Cutting-edge', or 'Leverage'.
# - Use lowercase occasionally. 
# - Use short, punchy sentences. 
# - Sound like you are sending a quick voice note text, not an email.
# """