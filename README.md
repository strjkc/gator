# This is Gator

Gator is an RSA feed aggregator.
It allows us to register users and for the users to follow RSA feeds and list out the articles in feeds.

### Dependencies which you need to install before running gator:

- Postgresql
- GO Compiler

### Set up Gator:
1. create gator db (assuming the default setup where the user is postgres and there is no password):
`psql -U posgres -c create database gator`

2. Go to the directory gator is cloned to:
`cd ~/gator`

3. Create tables:
`goose -dir sql/schema postgresql postgres://postgres:postgres@localhost:5432/gator?sslmode=disable up`

4. Set up the config file:
`echo "{\"db_url\":\"postgres://postgres:postgres@localhost:5432/gator?ssl\",\"username\":\"\"}" > ~/gatorconfig.json`

5. Build Gator:
`go build -o gator .`

6. Register your first user :
`./gator <username> command`

7. Subscribe to some feed using the:
`./gator addfeed <heedname> <linktofeed>`

8. Start the aggregation service using the:
`./gator agg <interval>s`

In a separate terminal 
The interval should be in the "<num of seconds>s" format
The application itself can be used in another terminal.

### The following commands are available:
- *login <username>* - this will switch the username in the config, no authentication is performed
- *register <username>* - register a new user in the db
- *reset* - it will clear the db 
- *users* - display all users
- *agg <interval>s* - it will fetch the posts from the registered and followed links in the set interval
- *addfeed <feed name>* - <link to feed> - it will add and follow the link to an RSA feed
- *feeds* - list all added feeds
- *follow* - follow an already added feed(if it was added by another user)
- *unfollow <feed url>* - unfollow the feed, but it will not be removed from the db
- *browse <optional number>* - by default will display the latest 2 posts, if optional number is provided it will display that number of posts 
