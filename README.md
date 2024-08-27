# URL Shortener Service

This Go-based URL shortener service allows users to shorten long URLs, retrieve analytics, update, and delete shortened URLs. It supports custom aliases, and automatically deletes expired URLs after a specified TTL (Time to Live).

## Features
 - URL Shortening: Generate short URLs for long URLs with optional custom aliases and TTL.

 - Redirection: Redirect to the original long URL when accessing the short URL.

 - Analytics: Track access counts and access times for each short URL.
Update URL: Update the alias or TTL of a shortened URL
.
 - Delete URL: Remove a shortened URL.

- Auto-Expiration: The service checks every 10 seconds for any expired URLs based on their TTL and deletes them from the store automatically.

## How to Run
Install Go: Ensure Go is installed on your machine.

```
git clone https://github.com/rootxrishabh/goURLshortner
```

```
cd goURLshortner
```

```
go run main.go
```

## Examples

1. Create a short URL.
```
curl --location 'http://localhost:8080/shorten' \
--header 'Content-Type: application/json' \
--data '{"long_url": "https://www.google.com/", "custom_alias": "test1", "ttl_seconds": 100}'
```

2. Access the short URL
```
curl --location 'http://localhost:8080/test1'
```

3. Get analytics of the short URL
```
curl --location 'http://localhost:8080/analytics/test1'
```

4. Update the short URL
```
curl --location 'http://localhost:8080/update/test1' \
--header 'Content-Type: application/json' \
--data '{"custom_alias": "test2", "ttl_seconds": 200}'
```

5. Delete the short URL
```
curl --location 'http://localhost:8080/delete/test1'
```

Access the Service: The server will start on http://localhost:8080.

## License

This project is licensed under the MIT License.
