### Accounts explorator

A bot that monitors some instances public feeds to grab all users seen and traverse their relationships.  

Feed a postgresql db with data collected : accounts, instances, followers, following...  

**NB**:  
_This is an experimental software that I use to discover how Mastodon network works._

#### TODO :
 * [] manage redirection when discovering/counting instances
 * [] monitor all public feeds of instances in db, either by websocket or pubsubhubbub
 * [] handle http/ws errors
 * [] remove accounts from db if a not found error was returned
 * [] fix stream error: stream ID 1; REFUSED_STREAM coming from nginx proxy
 * [] compute and send daily stats
 * [] automatic registration on new instance discovered
 * [] get followers/followings from puublic api (without registration), as a workaround to explore closed instances.
 
