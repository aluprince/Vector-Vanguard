import sqlite3
import json
import logging

DB_NAME = 'leads_vanguard.db'

def init_db():
    with sqlite3.connect(DB_NAME) as conn:
        cursor = conn.cursor()
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS leads (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                business_name TEXT,
                phone TEXT UNIQUE,
                audit_json TEXT,
                pitch TEXT,
                status TEXT DEFAULT 'pending',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        conn.commit()

def save_lead(lead_data):
    """Saves a lead if the phone number doesn't exist."""
    try:
        with sqlite3.connect(DB_NAME) as conn:
            cursor = conn.cursor()
            cursor.execute('''
                INSERT OR IGNORE INTO leads (business_name, phone, audit_json, pitch)
                VALUES (?, ?, ?, ?)
            ''', (
                lead_data['business_name'], 
                lead_data['phone'], 
                json.dumps(lead_data['audit']), 
                lead_data['pitch']
            ))
            conn.commit()
            return cursor.rowcount > 0 # Returns True if a new row was actually inserted
    except Exception as e:
        logging.error(f"Database Save Error: {e}")
        return False

def get_next_pending_lead():
    """Fetches the oldest pending lead."""
    with sqlite3.connect(DB_NAME) as conn:
        conn.row_factory = sqlite3.Row # Allows accessing columns by name
        cursor = conn.cursor()
        cursor.execute("SELECT * FROM leads WHERE status = 'pending' ORDER BY id ASC LIMIT 1")
        return cursor.fetchone()

def mark_as_sent(lead_id):
    """Updates status to sent."""
    with sqlite3.connect(DB_NAME) as conn:
        cursor = conn.cursor()
        cursor.execute("UPDATE leads SET status = 'sent' WHERE id = ?", (lead_id,))
        conn.commit()

def get_lead_by_id(lead_id):
    with sqlite3.connect(DB_NAME) as conn:
        conn.row_factory = sqlite3.Row
        cursor = conn.cursor()
        cursor.execute("SELECT * FROM leads WHERE id = ?", (lead_id,))
        return cursor.fetchone()