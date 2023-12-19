
import datetime


def log_message(text):
    timestamp = datetime.datetime.now().strftime("[%H:%M:%S]")
    log_text = f"{timestamp} {text}"
    return log_text