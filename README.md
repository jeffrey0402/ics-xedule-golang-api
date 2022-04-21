# ics-xedule-golang-api

A simple golang API that reads .ics feeds from xedule.

## To use: 

Get .ics feed URL, and add the following line to a .env file:
FEED_URL=url

Then, run the project.
Runs on port 8080 by default. 

The api first downloads the ics feed and stores it locally. Then it reads the values and stores it in RAM, to limit file accesses.

## Available endpoints:

`/rooster/classname`

GET request that gets all available rosters, where classname is the class code (Attendee CN in .ics file)

`/classes/`

GET all classes in the ics file. Assumes classnames contain an "\_" in their name, and teachers do not.

