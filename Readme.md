## tiktok-scraper-chrome-headless-go

TikTok's user profile scraper, if Captcha is detected, you have to resolve it manually, and press enter in terminal to continue.

### Basic use
```shell
./tiktok-scraper-linux-amd64 -profile-url https://www.tiktok.com/@username_here
```

### With video duration
```shell
./tiktok-scraper-linux-amd64 -profile-url https://www.tiktok.com/@username_here -video-duration
```

### If you have problem starting the scraper
```shell
./tiktok-scraper-linux-amd64 -profile-url https://www.tiktok.com/@username_here -disable-gpu
```

### All arguments
```shell
Usage of ./main:
  -debug-log
        Enable debug logging
  -disable-gpu
        Disable GPU
  -exec-path string
        Path to Chrome/Chromium or Brave executable
  -headless
        Run browser in headless mode
  -max-page-wait-sec int
        Maximum time to wait for page to load (default 20)
  -output string
        Output file (default "output.json")
  -profile-url string
        URL to the profile to scrape, e.g. https://www.tiktok.com/@username
  -video-duration
        Get video duration
```

### Example Result
```json
{
  "ProfileUrl": "https://www.tiktok.com/@xxxx",
  "StartedAt": "0001-01-01T00:00:00Z",
  "VideoStats": [
    {
      "Url": "https://www.tiktok.com/@xxxx/video/123412312?is_copy_url=1\u0026is_from_webapp=v1",
      "Description": "Foo bar",
      "Likes": "6.9K",
      "Comments": "699",
      "UploadAt": "11h ago",
      "DurationSeconds": 194.466667,
      "Date": "0001-01-01T00:00:00Z"
    },
    {
      "Url": "https://www.tiktok.com/@xxxx/video/123412312?is_copy_url=1\u0026is_from_webapp=v1",
      "Description": "Foo bar 2",
      "Likes": "96K",
      "Comments": "699",
      "UploadAt": "11h ago",
      "DurationSeconds": 69.69,
      "Date": "0001-01-01T00:00:00Z"
    }
  ]
}
```