import os
from typing import List, Optional, Dict
from datetime import datetime
import anthropic
import argparse
import yaml
import json


API_KEY = os.environ.get("CLAUDE_API_KEY")

class SourceImg:
    url: str
    filename: str
    caption: str


class SourceData:
    uuid: str
    content: List[str]
    images: List[SourceImg]
    title: str
    subtitle: Optional[str]


def build_system(topik_level: int) -> str:
    topik_word = ["beginner","intermediate","advanced"][(topik_level - 1) // 2]
    system = f"You are a language tutor for {topik_word} non-native Korean language students. Unless a different language is explicitly requested, you should use only idiomatic, natural sounding, and grammatically correct written Korean."
    return system


def build_prompt(topik_level: int, article: str) -> str:
    get_lessons = "Pull out several examples that could serve as the centerpiece of short language lessons. They could be grammatical points or related to typical written Korean style or rhetoric, but they should not be vocabulary focused. Give a short explanation for each lesson and illustrate it with 2-3 example sentences, one of which should be a direct quote from the article."
    get_context = "Briefly explain 2-3 pieces of background information that would help students understand this article. For example, you might explain historic, social or cultural facts that might be unfamiliar to non Koreans."
    get_questions = "Provide 3 comprehension questions to gauge students' understanding. These questions should each require 2-3 sentences to answer fully. For each question, also provide a sample answer."
    get_summary = "Provide a 2-3 sentence summary. It should generally avoid using specific names unless they refer to the subject of the article."
    get_tags = "Provide 3-4 news content categorization tags that could be used to find similar content in the same or adjacent categories."
    get_level = f"What TOPIK level would you rate this content? Please provide an explanation of why it is or is not appropriate for level {topik_level} students."
    get_vocabulary = f"From this article extract a list of vocabulary words that might be unfamiliar to TOPIK level {topik_level} students. For each word, give its Korean definition, an English translation, and 2-3 example sentences. One of the sample sentences should be pulled directly from the article."

    return_format = '''{
        "level": {"number": <int between 1 and 6>,
        "explanation": <str text explanation>},
        "lessons": [<list of objects>: {"title": str, "text": str}],
        "context": str,
        "questions": [<list of objects>: {"question": str, "answer": str}],
        "summary": str,
        "vocabulary": [<list of objects>: {"word": str, "definition": str, "english": str, "examples": [str]}],
        "tags": [str]
    }'''

    full_prompt = f"""
    You are creating content for TOPIK level {topik_level} students using the following source article:
    <article>{article}</article

    Provide answers to the following questions:
    1. Level: {get_level}
    2. Lessons: {get_lessons}
    3. Context: {get_context}
    4. Questions: {get_questions}
    5. Summary: {get_summary}
    6. Vocabulary: {get_vocabulary}
    7. Tags: {get_tags}

    Respond with ONLY a JSON object, using this format:
    {return_format}
    """
    return full_prompt

def get_basic(data: dict):
    client = anthropic.Anthropic()

    topik_level = 5
    article = data.get("content")
    uuid = data.get("short_uuid")
    if article is None or uuid is None:
        raise Exception(f"data cannot be processed: {data}")

    message = client.messages.create(
        model="claude-3-5-sonnet-20241022",
        max_tokens=8192,
        temperature=1,
        system=build_system(topik_level),
        messages=[
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": build_prompt(topik_level, "\n".join(article))
                    }
                ]
            }
        ]
    )
    try:
        out_data = json.loads(message.content[0].text)
    except Exception as e:
        print("Could not load message content\n", e)
        print(message.content)

    with open(f"content/{uuid}.yaml", "w", encoding="utf-8") as f:
        data.update(out_data)
        yaml.dump(data, f, encoding="utf-8", allow_unicode=True)
        print(f"Data written to {uuid}.yaml")


def get_simplified(data: dict, level: int):
    """
    replaces content with edited content for simplification
    preserve original uuid for tracking but flag as edited adn give it new uuid
    set the TOPIK level
    """
    return data


def create(uuid: str):
    data = None
    with open(f"content/{uuid}.yaml", "r") as f:
        data = yaml.safe_load(f)
    print("Data loaded\n", data)
    if data is not None:
        get_basic(data)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        prog='Daebak Create',
        description='Assistant to create new content',
        epilog='Now get busy creating something!')
    parser.add_argument("uuid", type=str)
    args = parser.parse_args()
    create(args.uuid)

