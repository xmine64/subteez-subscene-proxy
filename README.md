# Subteez-Subscene Proxy Server
This is a server for Subteez that proxies content from subscene.com
and provides them to you using Subteez API, so you can use it with
[Subteez App](https://play.google.com/store/apps/details?id=madamin.subtitles).

It also proxies banners and subtitle files, so people with censored
Internet (like Iranians) can use it without any problem.


## Using Subteez API
Send POST requests to endpoints, with parameters as json in body.
Responses are in json too.

```
endpoint      | description
--------      | -----------
/api/search   | Search for movies or series
/api/details  | Get details and subtitles files available for a movie or series
/api/download | Download subtitle file
```

You can check data structures that used for each request and response, [here](subteez/types.go).

## Using custom server with Subteez App
Go to settings and long press on Server, then you can enter your server address.

## License
All server codes are available to you with [MIT License](LICENSE), except all files in [static](static/) folder.
