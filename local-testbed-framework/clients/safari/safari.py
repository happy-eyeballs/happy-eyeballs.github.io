from selenium import webdriver
import sys


def main():
    if len(sys.argv) == 0:
        print("URL not specified as command line argument!", file=sys.stderr)
        exit(-1)

    url = sys.argv[1]

    driver = webdriver.Safari()
    if not driver:
        print("Unable to establish selenium session with Safari!", file=sys.stderr)
        exit(-1)

    driver.get(url)
    driver.quit()


if __name__ == "__main__":
    main()
