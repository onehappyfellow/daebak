from urllib.request import urlopen
import re
from bs4 import BeautifulSoup
import argparse
import yaml
from uuid import uuid4
from datetime import datetime
from claude import get_basic
import os
import requests


def parse_maeil(html):
    # content = html[html.find('class="article_content"'):]
    # content = re.findall("<p>.*?</p>", content)
    # title_results = re.search("<title.*?>.*?</title.*?>", html, re.IGNORECASE)
    # title = title_results.group()
    # title = re.sub("<.*?>", "", title)
    # minified_html = re.sub("\\n", "", html)
    # subtitle_results = re.search('<p class="subtitle".*?</p>', minified_html, re.IGNORECASE)
    # subtitle = subtitle_results.group()
    # subtitle = re.sub("<.*?>", "", subtitle).strip()

    soup = BeautifulSoup(html, "html.parser")
    data = {
        "doctype": "article",
        "source": "Maeil Shinmun",
        "title": str(soup.title.string),
        "content": [],
        "images": [],
    }
    for meta in soup.find_all("meta"):
        if meta.attrs.get("name") == "author":
            author = meta.attrs.get("content", "")
            data["author"] = re.sub("\s[\w\.]*?@\w*.com", "", author)
        if meta.attrs.get("name") == "keywords":
            data["keywords"] = meta.attrs.get("content").split(",")
    if data.get("author") is None:
        # try to find it an alternative way using CSS selectors
        try:
            data["author"] = soup.select_one("div.footer_byline a").text
        except AttributeError:
            pass

    body = soup.find(class_="article_content")
    for f in body.find_all("figure"):
        src = f.select_one("img")["src"]
        try:
            caption = f.select_one("figcaption").text.strip()
        except:
            caption = ""

        # download image
        local_src = download_image(src, uuid4())

        data["images"].append({"src": src, "local_src": local_src, "caption": caption})
    for p in body.find_all("p"):
        if "subtitle" in p.attrs.get("class", []):
            data["subtitle"] = p.text.strip()
        elif p.find("figure") is None:
            data["content"].append(p.text.strip())
    return data

def image_name_from_caption(uuid, caption):
    string = f"{uuid}-{caption.lower().replace(' ', '-')}"
    return string[:20].strip('-')

def download_image(url, filename, folder="content/images"):
    # Create the folder if it doesn't exist
    if not os.path.exists(folder):
        os.makedirs(folder)
        print(f"Created directory: {folder}")

    extension = url.split('.')[-1]
    if extension not in {"jpg", "jpeg", "png"}:
        print(f"Unknown extension: {extension}")
        extension = "jpg"

    filename = f"{filename}.{extension}"
    filepath = os.path.join(folder, filename)
    
    # Download the image
    try:
        response = requests.get(url, stream=True)
        response.raise_for_status()  # Raise an exception for HTTP errors
        
        # Save the image
        with open(filepath, 'wb') as file:
            for chunk in response.iter_content(chunk_size=8192):
                file.write(chunk)
        
        print(f"Image downloaded successfully to {filepath}")
        return filepath
    
    except requests.exceptions.RequestException as e:
        print(f"Error downloading image: {e}")
        return None

def scrape(url):
    page = urlopen(url)
    html = page.read().decode("utf-8")

    if "imaeil.com" in url:
        data = parse_maeil(html)
        print(f"Scraped and parsed page: {url}")
    else:
        print(f"Not Implemented: cannot parse page for {url}\n")
        print(html)
        print(f"Not Implemented: cannot parse page for {url}\n")
        return

    data["access_url"] = url
    data["access_date"] = datetime.now().isoformat()
    return data


def create(url):
    short_uuid = str(uuid4()).split("-")[0]
    data = scrape(url)
    data["short_uuid"] = short_uuid
    try:
        get_basic(data)
    except Exception as e:
        print(f"Error getting help from Claude:", e)
        with open(f"content/{short_uuid}_partial.yaml", "w", encoding="utf-8") as f:
            yaml.dump(data, f, encoding="utf-8", allow_unicode=True)
        print(f"Scraped data written to content/{short_uuid}_partial.yaml")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        prog='Daebak Create',
        description='Assistant to create new content',
        epilog='Now get busy creating something!')
    parser.add_argument("url", type=str)
    args = parser.parse_args()
    create(args.url)


