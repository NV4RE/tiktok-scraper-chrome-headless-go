package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
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
	VideoDate        = `[data-e2e="browser-nickname"] :nth-child(2)`
	VideoPlayer      = `[data-e2e="browse-video"] video`
	CaptchaVerify    = `!!document.querySelector('.captcha_verify_container')`
)

type Result struct {
	ProfileUrl string
	StartedAt  time.Time
	VideoStats []VideoStat
}

type VideoStat struct {
	Url             string
	Description     string
	Likes           string
	Comments        string
	UploadAt        string
	DurationSeconds float64
	Date            time.Time
}

func main() {
	execPath := flag.String("exec-path", "", "Path to Chrome/Chromium or Brave executable")
	headless := flag.Bool("headless", false, "Run browser in headless mode")
	profileUrl := flag.String("profile-url", "https://www.tiktok.com/@bimratcha", "URL to the profile to scrape, e.g. https://www.tiktok.com/@username")
	maxPageWaitSec := flag.Int("max-page-wait-sec", 20, "Maximum time to wait for page to load")
	debugLog := flag.Bool("debug-log", false, "Enable debug logging")
	output := flag.String("output", "output.json", "Output file")
	disableGpu := flag.Bool("disable-gpu", false, "Disable GPU")
	videoDuration := flag.Bool("video-duration", false, "Get video duration")
	flag.Parse()

	dir, err := os.MkdirTemp("", "chromedp-example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,

		// After Puppeteer's default behavior.
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),

		chromedp.UserDataDir(dir),
		chromedp.Flag("mute-audio", true),
	}

	if *execPath != "" {
		opts = append(opts, chromedp.ExecPath(*execPath))
	}

	if *headless {
		opts = append(opts, chromedp.Headless)
	}

	if *disableGpu {
		opts = append(opts, chromedp.DisableGPU)
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

	err = chromedp.Run(ctx,
		chromedp.Navigate(*profileUrl),
		chromedp.WaitVisible(VideosItem),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// wait for captcha
			time.Sleep(3 * time.Second)

			// check if captcha is present
			return checkCaptcha(ctx)
		}),
		chromedp.Click(VideosItem, chromedp.NodeVisible),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for {
				err := checkCaptcha(ctx)
				if err != nil {
					return err
				}

				vStat, err := getVideoStat(ctx, time.Duration(*maxPageWaitSec)*time.Second, *videoDuration)
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

func checkCaptcha(ctx context.Context) error {
	var captchaPresent bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(CaptchaVerify, &captchaPresent)); err != nil {
		return err
	}

	if captchaPresent {
		fmt.Println("Please solve the captcha and press enter in terminal")
		bufio.NewReader(os.Stdin).ReadBytes('\n')

		// in-case the captcha is not solved
		return checkCaptcha(ctx)
	}

	return nil
}

func getVideoStat(ctx context.Context, maxPageWait time.Duration, videoDuration bool) (VideoStat, error) {
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

	err = chromedp.Text(VideoDate, &vStat.UploadAt, chromedp.NodeVisible).Do(ctx)
	if err != nil {
		return vStat, err
	}

	if videoDuration {
		err := chromedp.Query(VideoPlayer,
			chromedp.WaitFunc(func(ctx context.Context, cur *cdp.Frame, rId runtime.ExecutionContextID, ids ...cdp.NodeID) ([]*cdp.Node, error) {
				if len(ids) == 0 {

					log.Println("len(ids) == 0")
					return nil, nil
				}

				var isVideoReady bool
				if err := chromedp.Evaluate(
					fmt.Sprintf(`document.querySelector('%s').readyState === 4`, VideoPlayer),
					&isVideoReady,
				).Do(ctx); err != nil {
					return nil, err
				}

				if isVideoReady {
					log.Println("video is ready")
					return []*cdp.Node{}, nil
				}

				log.Println("video not ready")
				return nil, nil

			}),
			chromedp.AtLeast(0),
		).Do(ctx)
		if err != nil {
			return vStat, err
		}

		err = chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s').duration`, VideoPlayer), &vStat.DurationSeconds).Do(ctx)
		if err != nil {
			return vStat, err
		}
	}

	log.Printf("Video stat: %+v", vStat)

	return vStat, err
}
