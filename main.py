import re
from datetime import datetime, timedelta
import pytz 

from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options


def config_driver():
    chrome_options = Options()
    # can be used for index, not articles?
    # chrome_options.page_load_strategy = 'none'
    driver = webdriver.Remote(
        command_executor='http://0.0.0.0:4444/wd/hub',
        options=chrome_options
    )
    driver.implicitly_wait(2)
    return driver

def chosun_index(driver, section: str):
    url = f"https://www.chosun.com/{section}/"
    print(">>> chosun_index", url)
    driver.get(url)
    print(">>> chosun_index", "get is done")
    elements = driver.find_elements(By.CSS_SELECTOR, ".story-card__headline")
    found = {}
    for el in elements:
        url = el.get_attribute("href")
        headline = el.text
        if url_is_valid(url):
            found[url] = headline
    print(f">>> found {len(found.keys())} current articles")
    return [{"url": key, "headline": value} for key, value in found.items()]

def chosun_article(driver, url):
    print(">>> chosun_article", url)
    driver.get(url)
    print(">>> chosun_article", "get is done")
    author = driver.find_element(By.CSS_SELECTOR, ".article-byline__author")
    headline = driver.find_element(By.CSS_SELECTOR, ".article-header__headline")
    paragraphs = driver.find_elements(By.CSS_SELECTOR, ".article-body p")
    return {
        "author": author.text,
        "headline": headline.text,
        "content": [p.text for p in paragraphs]
    }

def url_is_valid(url: str) -> bool:
    """
    returns true for urls that contain the current date
    formatted as YYYY/M/D
    function assumes seoul timezone (run before 9am MST)
    """
    HOURS = 26

    # screen out links that contain advertising link attribution
    if "utm_medium" in url:
        return False

    # validate current date
    date_string = re.search(r'\d{4}/\d{1,2}/\d{1,2}', url)
    if date_string is not None:
        try:
            pub_date = datetime.strptime(date_string.group(), '%Y/%m/%d')
            pub_korea = pytz.timezone("Asia/Seoul").localize(pub_date)
            now_korea = datetime.now(pytz.timezone("Asia/Seoul"))
        except ValueError as e:
            print(e)
            return False
        return (now_korea - pub_korea) < timedelta(hours=HOURS)

    # url does not have date, default to invalid
    return False


if __name__ == "__main__":
    driver = config_driver()
    try:
        # articles = chosun_index(driver, "economy")
        # print(type(articles), len(articles), articles)
        # sample = chosun_article(driver, articles[0].get("url"))
        sample = chosun_article(driver, "https://www.chosun.com/economy/tech_it/2024/11/01/VDWTAOUZ3ZDGHGZD4NQGHWJMYE/")
        print(type(sample), sample)
    finally:
        print(">>> driver.quit()")
        driver.quit()

