from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
import unittest

class SeleniumTest(unittest.TestCase):
    def setUp(self):
        # Configure Chrome options
        chrome_options = Options()
        chrome_options.page_load_strategy = 'none'
        
        # first start docker: docker run -d -p 4444:4444 -v /dev/shm:/dev/shm selenium/standalone-chrome -e SE_ENABLE_TRACING=false
        # Connect to remote Chrome instance
        self.driver = webdriver.Remote(
            command_executor='http://0.0.0.0:4444/wd/hub',
            options=chrome_options
        )
        self.driver.implicitly_wait(2)
    
    def tearDown(self):
        if self.driver:
            self.driver.quit()
    
    def test_website(self):
        # Navigate to the website
        self.driver.get('http://www.kenst.com/about/')
        
        # Assert the title matches expected value
        self.assertEqual(self.driver.title, "About")
        
        # Take a screenshot
        self.driver.save_screenshot('docker_image.png')

    def test_daebak(self):
        url = "https://www.chosun.com/economy/science/2024/10/31/W3JBXBEKDJAYJPCCGSQ6SWGBLE/"
        # Navigate to the website
        self.driver.get(url)

        print(">>> DEBUG page title:", self.driver.title)
        author = self.driver.find_element(By.CSS_SELECTOR, ".article-byline__author")
        print(">>> DEBUG author", author.text)
        headline = self.driver.find_element(By.CSS_SELECTOR, ".article-header__headline")
        print(">>> DEBUG headline", headline.text)

        paragraphs = self.driver.find_elements(By.CSS_SELECTOR, ".article-body p")
        text = ""
        for p in paragraphs:
            text += p.text
        print(">>> DEBUG paragraphs", text)
        
        # Take a screenshot
        self.driver.save_screenshot('docker_image2.png')

        self.assertTrue(False)

if __name__ == "__main__":
    unittest.main()