# Daebak Core

## Article to CMS
1) go run scraper.go

- copy title
- translate title, then `.lower().replace(" ","-")` for slug
- uuid is `from uuid import uuid4` then `uuid = str(uuid4()).split('-')[0]
  not guaranteed to be unique, so check against existing


- how do I determine the fitness of an article, for example screening out ones that are just too short?


## RUN
`python scrape.py <url>`


## TODO
- summary_en
- hanja
- comprehension questions 5
- block quote support
- article sub heading support
- author(s)