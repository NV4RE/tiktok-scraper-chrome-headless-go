package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"time"
)

const (
	VideosItem       = `[data-e2e="user-post-item"]`
	VideoDescription = `[data-e2e="browse-video-desc"]`
	VideoLikes       = `[data-e2e="browse-like-count"]`
	VideoComments    = `[data-e2e="browse-comment-count"]`
	VideoNextButton  = `[data-e2e="arrow-right"]`
	CaptchaVerify    = `!!document.querySelector('.captcha_verify_container')`
)

type Result struct {
	ProfileUrl string
	VideoStats []VideoStat
}

type VideoStat struct {
	Url         string
	Description string
	Likes       string
	Comments    string
}

func main() {
	execPath := flag.String("exec-path", "", "Path to Chrome/Chromium or Brave executable")
	headless := flag.Bool("headless", false, "Run browser in headless mode")
	profileUrl := flag.String("profile-url", "https://www.tiktok.com/@bimratcha", "URL to the profile to scrape, e.g. https://www.tiktok.com/@username")
	maxPageWaitSec := flag.Int("max-page-wait-sec", 20, "Maximum time to wait for page to load")
	debugLog := flag.Bool("debug-log", false, "Enable debug logging")
	output := flag.String("output", "output.json", "Output file")
	flag.Parse()

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserDataDir(""),
		chromedp.Flag("new-instance", true),
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("mute-audio", true),
	}

	if *execPath != "" {
		opts = append(opts, chromedp.ExecPath(*execPath))
	}

	if *headless {
		opts = append(opts, chromedp.Headless)
	}

	ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	var ctxOpts []chromedp.ContextOption

	if *debugLog {
		ctxOpts = append(ctxOpts, chromedp.WithLogf(log.Printf))
	}

	ctx, cancel := chromedp.NewContext(ctx, ctxOpts...)
	defer cancel()

	result := Result{
		ProfileUrl: *profileUrl,
		VideoStats: make([]VideoStat, 0),
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(*profileUrl),
		chromedp.WaitVisible(VideosItem),
		chromedp.Click(VideosItem, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// wait for captcha
			time.Sleep(3 * time.Second)

			// check if captcha is present
			var captchaPresent bool
			if err := chromedp.Run(ctx, chromedp.Evaluate(CaptchaVerify, &captchaPresent)); err != nil {
				return err
			}

			// if captcha is present, wait for it to be solved, and press enter in terminal
			if captchaPresent {
				fmt.Println("Please solve the captcha and press enter in terminal")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
			}

			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for {
				vStat, err := getVideoStat(ctx, time.Duration(*maxPageWaitSec)*time.Second)
				if err != nil {
					return err
				}

				result.VideoStats = append(result.VideoStats, vStat)

				// write to file
				b, err := json.Marshal(result)
				if err != nil {
					return err
				}

				err = os.WriteFile(*output, b, 0644)
				if err != nil {
					return err
				}

				err = chromedp.Click(VideoNextButton, chromedp.NodeVisible).Do(ctx)
				if err != nil {
					return err
				}
			}
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func getVideoStat(ctx context.Context, maxPageWait time.Duration) (VideoStat, error) {
	vStat := VideoStat{}
	ctx, cancel := context.WithTimeout(ctx, maxPageWait)
	defer cancel()

	err := chromedp.Location(&vStat.Url).Do(ctx)
	if err != nil {
		return vStat, err
	}

	err = chromedp.Text(VideoDescription, &vStat.Description, chromedp.NodeVisible).Do(ctx)
	if err != nil {
		return vStat, err
	}

	err = chromedp.Text(VideoLikes, &vStat.Likes, chromedp.NodeVisible).Do(ctx)
	if err != nil {
		return vStat, err
	}

	err = chromedp.Text(VideoComments, &vStat.Comments, chromedp.NodeVisible).Do(ctx)
	if err != nil {
		return vStat, err
	}

	log.Printf("URL: %s, Description: %s, Likes: %s, Comments: %s", vStat.Url, vStat.Description, vStat.Likes, vStat.Comments)

	return vStat, err
}
