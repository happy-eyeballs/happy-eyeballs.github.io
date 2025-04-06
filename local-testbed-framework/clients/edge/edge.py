from selenium import webdriver
import sys
import time

from selenium.webdriver.edge.options import Options as EdgeOptions

DOCKER_SELENIUM_URL = 'http://localhost:4444'


def main():
    if len(sys.argv) == 0:
        print("URL not specified as command line argument!", file=sys.stderr)
        exit(-1)

    url = sys.argv[1]

    driver = establish_selenium_connection(60, 1)
    if not driver:
        print("Unable to establish selenium session with remote browser!", file=sys.stderr)
        exit(-1)

    driver.get(url)
    driver.quit()


def establish_selenium_connection(retries: int, retry_delay: float):
    for retry in range(retries):
        try:
            driver = webdriver.Remote(command_executor=DOCKER_SELENIUM_URL, options=EdgeOptions())
            return driver
        except (Exception,):
            time.sleep(retry_delay)
            continue

    return None


if __name__ == "__main__":
    main()
