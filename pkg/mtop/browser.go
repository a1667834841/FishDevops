package mtop

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/playwright-community/playwright-go"
)

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	Headless bool // 是否无头模式
}

// CookieResult Cookie 获取结果
type CookieResult struct {
	Token   string
	Cookies []*http.Cookie
}

// GetCookiesWithBrowser 使用浏览器获取闲鱼 Cookie
func GetCookiesWithBrowser(config BrowserConfig) (*CookieResult, error) {
	headless := config.Headless
	rand.Seed(time.Now().UnixNano())

	// 初始化 Playwright
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		return nil, fmt.Errorf("安装Playwright浏览器失败: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("启动Playwright失败: %w", err)
	}
	defer pw.Stop()

	// 随机选择 User-Agent
	userAgent := getRandomUserAgent()

	// 随机 viewport 尺寸
	viewportWidth := 1920 - rand.Intn(400)  // 1520-1920
	viewportHeight := 1080 - rand.Intn(300) // 780-1080

	// 启动浏览器 - 添加反检测参数
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
			"--disable-features=IsolateOrigins,site-per-process",
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-web-security",
			"--disable-features=VizDisplayCompositor",
			"--start-maximized",
			"--disable-infobars",
			"--window-position=0,0",
		},
		Channel: playwright.String("chrome"), // 使用系统安装的Chrome（如果可用）
	})
	if err != nil {
		return nil, fmt.Errorf("启动浏览器失败: %w", err)
	}
	defer browser.Close()

	// 创建浏览器上下文 - 设置更真实的浏览器参数
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent:        playwright.String(userAgent),
		Viewport:         &playwright.Size{Width: viewportWidth, Height: viewportHeight},
		Locale:           playwright.String("zh-CN"),
		TimezoneId:       playwright.String("Asia/Shanghai"),
		Permissions:      []string{"geolocation", "notifications"},
		Geolocation:      &playwright.Geolocation{Latitude: 31.2304, Longitude: 121.4737}, // 上海
		ColorScheme:      playwright.ColorSchemeLight,
		DeviceScaleFactor: playwright.Float(1),
		HasTouch:         playwright.Bool(false),
		IsMobile:         playwright.Bool(false),
		AcceptDownloads:  playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
		BypassCSP:        playwright.Bool(true),
		JavaScriptEnabled: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("创建上下文失败: %w", err)
	}
	defer context.Close()

	// 添加初始化脚本 - 注入反检测代码
	err = context.AddInitScript(playwright.Script{Content: playwright.String(getAntiDetectionScript())})
	if err != nil {
		return nil, fmt.Errorf("添加反检测脚本失败: %w", err)
	}

	// 创建新页面
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	// 设置额外的超时时间
	page.SetDefaultTimeout(60000)
	page.SetDefaultNavigationTimeout(60000)

	// 导航到闲鱼网站
	_, err = page.Goto("https://www.goofish.com", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return nil, fmt.Errorf("打开闲鱼网站失败: %w", err)
	}

	// 检查是否需要登录
	needLogin, _ := page.Evaluate("() => { return !!document.querySelector('.login-guide') || document.body.innerText.includes('立即登录') }")
	if needLogin != nil && needLogin.(bool) {
		// 非无头模式下提示用户登录
		if !headless {
			fmt.Println("\n========================================")
			fmt.Println("  请在浏览器中登录闲鱼账号")
			fmt.Println("========================================")
			fmt.Println("等待用户登录...")

			// 等待用户登录（最多等待5分钟）
			_, err := page.WaitForFunction("() => { return !document.querySelector('.login-guide') && !document.body.innerText.includes('立即登录') }", nil)
			if err != nil {
				return nil, fmt.Errorf("等待登录超时，请确保已登录闲鱼账号")
			}
			fmt.Println("✅ 检测到登录成功！")
			// 登录后再等待一下让 cookie 生成
			time.Sleep(2 * time.Second)
		} else {
			return nil, fmt.Errorf("检测到未登录，请先在浏览器中登录闲鱼账号，或使用 headless=false 模式运行")
		}
	}

	// 等待页面完全加载并执行JavaScript
	// 等待一小段时间让异步脚本执行
	time.Sleep(time.Duration(2000+rand.Intn(2000)) * time.Millisecond)

	// 等待token生成 - 尝试等待token cookie出现
	_, err = page.WaitForFunction("() => { return document.cookie.includes('_m_h5_tk') }", nil)
	if err != nil {
		// 如果等待失败，再等待一段时间作为后备
		time.Sleep(3 * time.Second)
	}

	// 获取 Cookies
	cookies, err := context.Cookies()
	if err != nil {
		return nil, fmt.Errorf("获取Cookies失败: %w", err)
	}

	// 检查关键 Cookie 是否存在
	hasCookie2 := false
	hasUnb := false
	for _, c := range cookies {
		if c.Name == "cookie2" && c.Value != "" {
			hasCookie2 = true
		}
		if c.Name == "unb" && c.Value != "" {
			hasUnb = true
		}
	}

	if !hasCookie2 || !hasUnb {
		fmt.Println("\n⚠️  警告: 检测到登录状态不完整")
		if !hasCookie2 {
			fmt.Println("   - 缺少 cookie2")
		}
		if !hasUnb {
			fmt.Println("   - 缺少 unb (用户ID)")
		}
		fmt.Println("   可能导致 API 调用失败")
	}

	// 转换为 http.Cookie 格式
	cookieMaps := make([]map[string]string, len(cookies))
	for i, c := range cookies {
		cookieMaps[i] = map[string]string{
			"name":  c.Name,
			"value": c.Value,
		}
		if c.Domain != "" {
			cookieMaps[i]["domain"] = c.Domain
		}
		if c.Path != "" {
			cookieMaps[i]["path"] = c.Path
		}
	}

	httpCookies := ConvertMapSliceToHTTPCookies(cookieMaps)
	token := GetTokenFromCookies(httpCookies)

	if token == "" {
		return nil, fmt.Errorf("未获取到 Token，请确保已登录闲鱼")
	}

	return &CookieResult{
		Token:   token,
		Cookies: httpCookies,
	}, nil
}

// PrintStartupInfo 打印启动信息
func PrintStartupInfo(token string) {
	fmt.Println("🌐 闲鱼 API 服务")
	fmt.Println("=================")
	if token != "" {
		fmt.Printf("✅ 获取到 Token: %s...\n", token[:10])
	}
}

// getRandomUserAgent 随机获取一个真实的浏览器 User-Agent
func getRandomUserAgent() string {
	userAgents := []string{
		// Chrome on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		// Chrome on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		// Edge on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
		// Safari on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
		// Firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		// Chrome on Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
	return userAgents[rand.Intn(len(userAgents))]
}

// getAntiDetectionScript 获取反检测脚本，隐藏自动化特征
func getAntiDetectionScript() string {
	return `
		// 覆盖 navigator.webdriver 属性
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});

		// 覆盖 chrome 对象
		window.chrome = {
			runtime: {}
		};

		// 覆盖 permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);

		// 覆盖 plugins 长度
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5]
		});

		// 覆盖 languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['zh-CN', 'zh', 'en-US', 'en']
		});

		// 添加真实的 plugins
		Object.defineProperty(navigator, 'plugins', {
			get: () => {
				return {
					length: 3,
					0: {
						name: 'Chrome PDF Plugin',
						filename: 'internal-pdf-viewer',
						description: 'Portable Document Format'
					},
					1: {
						name: 'Chrome PDF Viewer',
						filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai',
						description: ''
					},
					2: {
						name: 'Native Client',
						filename: 'internal-nacl-plugin',
						description: ''
					}
				};
			}
		});

		// 覆盖连接信息
		Object.defineProperty(navigator, 'connection', {
			get: () => ({
				effectiveType: '4g',
				rtt: 50,
				downlink: 10
			})
		});

		// 覆盖 deviceMemory
		Object.defineProperty(navigator, 'deviceMemory', {
			get: () => 8
		});

		// 覆盖 hardwareConcurrency
		Object.defineProperty(navigator, 'hardwareConcurrency', {
			get: () => 8
		});

		// 覆盖 maxTouchPoints
		Object.defineProperty(navigator, 'maxTouchPoints', {
			get: () => 0
		});

		// 隐藏自动化相关属性
		delete navigator.__proto__.webdriver;

		// 覆盖外层高度
		Object.defineProperty(window, 'outerHeight', {
			get: () => window.innerHeight
		});

		// 覆盖外层宽度
		Object.defineProperty(window, 'outerWidth', {
			get: () => window.innerWidth
		});

		// 模拟真实的屏幕方向
		Object.defineProperty(screen, 'availWidth', {
			get: () => screen.width
		});
		Object.defineProperty(screen, 'availHeight', {
			get: () => screen.height - 40
		});

		// 覆盖 getParameter
		const originalGetParameter = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(parameter) {
			if (parameter === 37445) {
				return 'Intel Inc.';
			}
			if (parameter === 37446) {
				return 'Intel Iris OpenGL Engine';
			}
			return originalGetParameter(parameter);
		};

		// 添加真实的 window.navigator.platform
		Object.defineProperty(navigator, 'platform', {
			get: () => 'Win32'
		});

		// 防止检测
		window.addEventListener('devtoolschange', (event) => {
			event.preventDefault();
		});

		// 隐藏 Playwright/Headless 特征
		Object.defineProperty(navigator, 'headless', {
			get: () => undefined
		});
	`
}
