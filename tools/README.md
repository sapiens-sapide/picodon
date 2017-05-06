### Accounts explorator

A bot that monitors some instances public feeds to grab all users seen and traverse their relationships.  

Feed a postgresql db with data collected : accounts, instances, followers, following...  

**NB**:  
_This is an experimental software that I use to discover how Mastodon network works._

#### TODO :
 * [] Fetch the `/about/more` page from listed instances to count declared accounts.
 * [] handle http/ws errors
 * [] fix stream error: stream ID 1; REFUSED_STREAM coming from nginx proxy
 * [] compute and send daily stats
 * [] automatic registration on new instance discovered
 * [] get followers/followings from puublic api (without registration), as a workaround to explore closed instances.
 
