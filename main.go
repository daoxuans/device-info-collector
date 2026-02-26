package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Response 统一响应结构体
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DeviceInfo 结构体定义
type DeviceInfo struct {
	Timestamp           string `json:"timestamp"`
	UserAgent           string `json:"userAgent"`
	IPAddress           string `json:"ipAddress"`
	Screen              string `json:"screen"`
	ColorDepth          string `json:"colorDepth"`
	Timezone            string `json:"timezone"`
	Language            string `json:"language"`
	Platform            string `json:"platform"`
	CPUCores            string `json:"cpuCores"`
	DeviceMemory        string `json:"deviceMemory"`
	Connection          string `json:"connection"`
	TouchSupport        string `json:"touchSupport"`
	PixelRatio          string `json:"pixelRatio"`
	AvailableScreen     string `json:"availableScreen"`
	CookiesEnabled      string `json:"cookiesEnabled"`
	JavaEnabled         string `json:"javaEnabled"`
	DoNotTrack          string `json:"doNotTrack"`
	HardwareConcurrency string `json:"hardwareConcurrency"`
	Vendor              string `json:"vendor"`
	Product             string `json:"product"`
	// 新增字段
	Battery           string `json:"battery"`
	OnlineStatus      string `json:"onlineStatus"`
	MaxTouchPoints    string `json:"maxTouchPoints"`
	PDFViewer         string `json:"pdfViewer"`
	WebGL             string `json:"webgl"`
	Canvas            string `json:"canvas"`
	AudioContext      string `json:"audioContext"`
	LocalStorage      string `json:"localStorage"`
	SessionStorage    string `json:"sessionStorage"`
	IndexedDB         string `json:"indexedDB"`
	Geolocation       string `json:"geolocation"`
	LocationDetails   string `json:"locationDetails"`
	Notifications     string `json:"notifications"`
	ServiceWorker     string `json:"serviceWorker"`
	WebRTC            string `json:"webrtc"`
	MediaDevices      string `json:"mediaDevices"`
	DeviceOrientation string `json:"deviceOrientation"`
	Vibration         string `json:"vibration"`
	Bluetooth         string `json:"bluetooth"`
	USB               string `json:"usb"`
	Clipboard         string `json:"clipboard"`
	Share             string `json:"share"`
	PaymentRequest    string `json:"paymentRequest"`
	Accelerometer     string `json:"accelerometer"`
	Gyroscope         string `json:"gyroscope"`
	Magnetometer      string `json:"magnetometer"`
	GamepadAPI        string `json:"gamepadAPI"`
	VRDisplay         string `json:"vrDisplay"`
	WebAssembly       string `json:"webAssembly"`
	CSSFeatures       string `json:"cssFeatures"`
	FontList          string `json:"fontList"`
	Plugins           string `json:"plugins"`
	MimeTypes         string `json:"mimeTypes"`
	ViewportSize      string `json:"viewportSize"`
	DeviceType        string `json:"deviceType"`
	OSVersion         string `json:"osVersion"`
	BrowserVersion    string `json:"browserVersion"`
	ReferrerPolicy    string `json:"referrerPolicy"`
	HTTPSSupport      string `json:"httpsSupport"`
	// Canvas指纹相关
	CanvasFingerprint string `json:"canvasFingerprint"`
	WebGLFingerprint  string `json:"webglFingerprint"`
	FontFingerprint   string `json:"fontFingerprint"`
	// 音频指纹
	AudioFingerprint string `json:"audioFingerprint"`
	// 屏幕详细信息
	ScreenOrientation string `json:"screenOrientation"`
	ColorGamut        string `json:"colorGamut"`
	HDR               string `json:"hdr"`
	RefreshRate       string `json:"refreshRate"`
	// 网络详细信息
	DownlinkMax string `json:"downlinkMax"`
	NetworkType string `json:"networkType"`
	// 性能信息
	MemoryInfo       string `json:"memoryInfo"`
	NavigationTiming string `json:"navigationTiming"`
	// 键盘布局
	KeyboardLayout string `json:"keyboardLayout"`
	// 语言偏好
	Languages string `json:"languages"`
	// 媒体能力
	MediaCapabilities string `json:"mediaCapabilities"`
	VideoCodecs       string `json:"videoCodecs"`
	AudioCodecs       string `json:"audioCodecs"`
	// WebGL详细信息
	WebGLVendor   string `json:"webglVendor"`
	WebGLRenderer string `json:"webglRenderer"`
	// 安全与隐私
	AdBlocker   string `json:"adBlocker"`
	PrivateMode string `json:"privateMode"`
	// 时间信息
	TimezoneOffset string `json:"timezoneOffset"`
	SystemTime     string `json:"systemTime"`
	// Pointer能力
	PointerType  string `json:"pointerType"`
	HoverCapable string `json:"hoverCapable"`
	// 动画帧率
	AnimationFrameRate string `json:"animationFrameRate"`
}

// Rating 评分结构体
type Rating struct {
	Score int `json:"score"`
}

// RatingStats 评分统计
type RatingStats struct {
	Total   int     `json:"total"`
	Average float64 `json:"average"`
	Counts  [5]int  `json:"counts"`
}

// 内存评分存储
var (
	ratingsMutex sync.Mutex
	ratings      []int
)

// 限流器结构
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.Mutex
}

var rateLimiter = &RateLimiter{
	requests: make(map[string][]time.Time),
}

// 检查是否允许请求 (每分钟最多30次)
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	requests := rl.requests[ip]

	// 清理1分钟前的请求
	validRequests := make([]time.Time, 0)
	for _, req := range requests {
		if now.Sub(req) < time.Minute {
			validRequests = append(validRequests, req)
		}
	}

	if len(validRequests) >= 30 {
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	return true
}

// 获取客户端真实IP
func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// 发送JSON响应
func sendJSONResponse(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// 处理设备信息提交
func collectHandler(w http.ResponseWriter, r *http.Request) {
	// CORS预检请求
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		fmt.Printf("错误: 收到非POST请求, 方法: %s\n", r.Method)
		sendJSONResponse(w, http.StatusMethodNotAllowed, Response{
			Status:  "error",
			Message: "Only POST method is allowed",
		})
		return
	}

	// 限流检查
	ip := getClientIP(r)
	if !rateLimiter.Allow(ip) {
		fmt.Printf("限流: IP %s 请求过于频繁\n", ip)
		sendJSONResponse(w, http.StatusTooManyRequests, Response{
			Status:  "error",
			Message: "请求过于频繁，请稍后再试",
		})
		return
	}

	// 打印请求头信息用于调试
	fmt.Printf("收到请求 - IP: %s, Content-Type: %s, Content-Length: %s\n",
		ip, r.Header.Get("Content-Type"), r.Header.Get("Content-Length"))

	var info DeviceInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		fmt.Printf("JSON解析错误: %v\n", err)
		sendJSONResponse(w, http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	// 设置时间戳和IP地址
	info.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	info.IPAddress = ip

	// 控制台输出
	fmt.Printf("收集到设备信息 [%s] IP: %s, UserAgent: %s\n",
		info.Timestamp, info.IPAddress, info.UserAgent)

	// 返回成功响应
	sendJSONResponse(w, http.StatusOK, Response{
		Status:  "success",
		Message: "设备信息收集成功",
		Data:    info,
	})
}

// 提供前端页面
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>设备信息收集</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6; color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh; padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { text-align: center; margin-bottom: 30px; color: white; }
        .header h1 { font-size: 2.5rem; margin-bottom: 10px; text-shadow: 2px 2px 4px rgba(0,0,0,0.3); }
        .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(350px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .info-card {
            background: white; border-radius: 15px; padding: 25px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }
        .info-card:hover { transform: translateY(-5px); box-shadow: 0 15px 40px rgba(0,0,0,0.3); }
        .info-card h3 {
            color: #4a5568; border-bottom: 2px solid #e2e8f0; padding-bottom: 10px;
            margin-bottom: 15px; display: flex; align-items: center; gap: 10px;
        }
        .info-item {
            display: flex; justify-content: space-between; align-items: center;
            padding: 8px 0; border-bottom: 1px solid #f7fafc;
        }
        .info-item:last-child { border-bottom: none; }
        .info-label { font-weight: 600; color: #4a5568; flex-shrink: 0; }
        .info-value { color: #2d3748; text-align: right; word-break: break-all; flex: 1; margin-left: 15px; }
        .status {
            text-align: center; padding: 20px; background: white; border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2); margin-bottom: 20px;
        }
        .status.success { background: #48bb78; color: white; }
        .status.error { background: #f56565; color: white; }
        .actions { text-align: center; margin-top: 20px; }
        .btn {
            background: #4299e1; color: white; border: none; padding: 12px 30px;
            border-radius: 25px; cursor: pointer; font-size: 1rem;
            transition: all 0.3s ease; box-shadow: 0 4px 15px rgba(66, 153, 225, 0.3);
        }
        .btn:hover { background: #3182ce; transform: translateY(-2px); box-shadow: 0 6px 20px rgba(66, 153, 225, 0.4); }
        .rating-card {
            background: white; border-radius: 15px; padding: 25px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2); margin-bottom: 20px; text-align: center;
        }
        .rating-card h3 { color: #4a5568; margin-bottom: 15px; font-size: 1.3rem; }
        .stars { display: flex; justify-content: center; gap: 10px; margin: 15px 0; }
        .star {
            font-size: 2.5rem; cursor: pointer; transition: transform 0.2s ease, color 0.2s ease;
            color: #cbd5e0; line-height: 1;
        }
        .star:hover, .star.active { color: #f6ad55; transform: scale(1.2); }
        .rating-msg { margin-top: 10px; font-size: 0.95rem; color: #718096; min-height: 1.4em; }
        .rating-stats { margin-top: 15px; font-size: 0.9rem; color: #4a5568; }
        .rating-stats .avg { font-size: 2rem; font-weight: bold; color: #f6ad55; }
        .rating-bar-row { display: flex; align-items: center; gap: 8px; margin: 3px 0; font-size: 0.85rem; }
        .rating-bar-bg { flex: 1; background: #e2e8f0; border-radius: 4px; height: 10px; }
        .rating-bar-fill { height: 10px; border-radius: 4px; background: #f6ad55; transition: width 0.4s; }
        @media (max-width: 768px) {
            .info-grid { grid-template-columns: 1fr; }
            .header h1 { font-size: 2rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🖥️ 设备信息收集</h1>
            <p>以下展示了当前浏览器和设备可获取的信息</p>
        </div>
        
        <div id="status" class="status"></div>
        
        <div class="info-grid">
            <div class="info-card">
                <h3>🎯 设备指纹</h3>
                <div class="info-item"><span class="info-label">Canvas指纹:</span><span class="info-value" id="canvasFingerprint" style="font-family: monospace; font-size: 0.8em;">生成中...</span></div>
                <div class="info-item"><span class="info-label">WebGL指纹:</span><span class="info-value" id="webglFingerprint" style="font-family: monospace; font-size: 0.8em;">生成中...</span></div>
                <div class="info-item"><span class="info-label">字体指纹:</span><span class="info-value" id="fontFingerprint" style="font-family: monospace; font-size: 0.8em;">生成中...</span></div>
            </div>

            <div class="info-card">
                <h3>🌐 浏览器信息</h3>
                <div class="info-item"><span class="info-label">User Agent:</span><span class="info-value" id="userAgent">检测中...</span></div>
                <div class="info-item"><span class="info-label">平台:</span><span class="info-value" id="platform">检测中...</span></div>
                <div class="info-item"><span class="info-label">语言:</span><span class="info-value" id="language">检测中...</span></div>
                <div class="info-item"><span class="info-label">浏览器厂商:</span><span class="info-value" id="vendor">检测中...</span></div>
                <div class="info-item"><span class="info-label">浏览器产品:</span><span class="info-value" id="product">检测中...</span></div>
                <div class="info-item"><span class="info-label">浏览器版本:</span><span class="info-value" id="browserVersion">检测中...</span></div>
            </div>
            
            <div class="info-card">
                <h3>🖥️ 显示信息</h3>
                <div class="info-item"><span class="info-label">屏幕分辨率:</span><span class="info-value" id="screen">检测中...</span></div>
                <div class="info-item"><span class="info-label">可用屏幕:</span><span class="info-value" id="availableScreen">检测中...</span></div>
                <div class="info-item"><span class="info-label">视口大小:</span><span class="info-value" id="viewportSize">检测中...</span></div>
                <div class="info-item"><span class="info-label">颜色深度:</span><span class="info-value" id="colorDepth">检测中...</span></div>
                <div class="info-item"><span class="info-label">像素比:</span><span class="info-value" id="pixelRatio">检测中...</span></div>
            </div>
            
            <div class="info-card">
                <h3>⚙️ 系统信息</h3>
                <div class="info-item"><span class="info-label">设备类型:</span><span class="info-value" id="deviceType">检测中...</span></div>
                <div class="info-item"><span class="info-label">操作系统:</span><span class="info-value" id="osVersion">检测中...</span></div>
                <div class="info-item"><span class="info-label">时区:</span><span class="info-value" id="timezone">检测中...</span></div>
                <div class="info-item"><span class="info-label">CPU核心:</span><span class="info-value" id="cpuCores">检测中...</span></div>
                <div class="info-item"><span class="info-label">设备内存:</span><span class="info-value" id="deviceMemory">检测中...</span></div>
                <div class="info-item"><span class="info-label">硬件并发:</span><span class="info-value" id="hardwareConcurrency">检测中...</span></div>
            </div>
            
            <div class="info-card">
                <h3>📡 网络与连接</h3>
                <div class="info-item"><span class="info-label">连接类型:</span><span class="info-value" id="connection">检测中...</span></div>
                <div class="info-item"><span class="info-label">在线状态:</span><span class="info-value" id="onlineStatus">检测中...</span></div>
                <div class="info-item"><span class="info-label">HTTPS支持:</span><span class="info-value" id="httpsSupport">检测中...</span></div>
                <div class="info-item"><span class="info-label">IP地址:</span><span class="info-value" id="ipAddress">检测中...</span></div>
                <div class="info-item"><span class="info-label">时间戳:</span><span class="info-value" id="timestamp">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🔧 硬件功能</h3>
                <div class="info-item"><span class="info-label">触摸支持:</span><span class="info-value" id="touchSupport">检测中...</span></div>
                <div class="info-item"><span class="info-label">最大触点:</span><span class="info-value" id="maxTouchPoints">检测中...</span></div>
                <div class="info-item"><span class="info-label">电池状态:</span><span class="info-value" id="battery">检测中...</span></div>
                <div class="info-item"><span class="info-label">振动支持:</span><span class="info-value" id="vibration">检测中...</span></div>
                <div class="info-item"><span class="info-label">设备方向:</span><span class="info-value" id="deviceOrientation">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🎮 传感器与游戏</h3>
                <div class="info-item"><span class="info-label">加速度计:</span><span class="info-value" id="accelerometer">检测中...</span></div>
                <div class="info-item"><span class="info-label">陀螺仪:</span><span class="info-value" id="gyroscope">检测中...</span></div>
                <div class="info-item"><span class="info-label">磁力计:</span><span class="info-value" id="magnetometer">检测中...</span></div>
                <div class="info-item"><span class="info-label">游戏手柄:</span><span class="info-value" id="gamepadAPI">检测中...</span></div>
                <div class="info-item"><span class="info-label">VR显示:</span><span class="info-value" id="vrDisplay">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🎨 图形与媒体</h3>
                <div class="info-item"><span class="info-label">WebGL:</span><span class="info-value" id="webgl">检测中...</span></div>
                <div class="info-item"><span class="info-label">Canvas:</span><span class="info-value" id="canvas">检测中...</span></div>
                <div class="info-item"><span class="info-label">音频上下文:</span><span class="info-value" id="audioContext">检测中...</span></div>
                <div class="info-item"><span class="info-label">媒体设备:</span><span class="info-value" id="mediaDevices">检测中...</span></div>
                <div class="info-item"><span class="info-label">WebRTC:</span><span class="info-value" id="webrtc">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>💾 存储与API</h3>
                <div class="info-item"><span class="info-label">本地存储:</span><span class="info-value" id="localStorage">检测中...</span></div>
                <div class="info-item"><span class="info-label">会话存储:</span><span class="info-value" id="sessionStorage">检测中...</span></div>
                <div class="info-item"><span class="info-label">IndexedDB:</span><span class="info-value" id="indexedDB">检测中...</span></div>
                <div class="info-item"><span class="info-label">Service Worker:</span><span class="info-value" id="serviceWorker">检测中...</span></div>
                <div class="info-item"><span class="info-label">WebAssembly:</span><span class="info-value" id="webAssembly">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🔐 权限与隐私</h3>
                <div class="info-item"><span class="info-label">Cookie支持:</span><span class="info-value" id="cookiesEnabled">检测中...</span></div>
                <div class="info-item"><span class="info-label">Do Not Track:</span><span class="info-value" id="doNotTrack">检测中...</span></div>
                <div class="info-item"><span class="info-label">地理位置:</span><span class="info-value" id="geolocation">检测中...</span></div>
                <div class="info-item"><span class="info-label">位置详情:</span><span class="info-value" id="locationDetails">获取中...</span></div>
                <div class="info-item"><span class="info-label">通知权限:</span><span class="info-value" id="notifications">检测中...</span></div>
                <div class="info-item"><span class="info-label">剪贴板:</span><span class="info-value" id="clipboard">检测中...</span></div>
                <div class="info-item"><span class="info-label">广告拦截:</span><span class="info-value" id="adBlocker">检测中...</span></div>
                <div class="info-item"><span class="info-label">隐私模式:</span><span class="info-value" id="privateMode">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🎨 屏幕与显示</h3>
                <div class="info-item"><span class="info-label">屏幕方向:</span><span class="info-value" id="screenOrientation">检测中...</span></div>
                <div class="info-item"><span class="info-label">色域:</span><span class="info-value" id="colorGamut">检测中...</span></div>
                <div class="info-item"><span class="info-label">HDR支持:</span><span class="info-value" id="hdr">检测中...</span></div>
                <div class="info-item"><span class="info-label">刷新率:</span><span class="info-value" id="refreshRate">检测中...</span></div>
                <div class="info-item"><span class="info-label">动画帧率:</span><span class="info-value" id="animationFrameRate">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🌐 网络详情</h3>
                <div class="info-item"><span class="info-label">网络类型:</span><span class="info-value" id="networkType">检测中...</span></div>
                <div class="info-item"><span class="info-label">最大下行速度:</span><span class="info-value" id="downlinkMax">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>⚡ 性能信息</h3>
                <div class="info-item"><span class="info-label">内存使用:</span><span class="info-value" id="memoryInfo">检测中...</span></div>
                <div class="info-item"><span class="info-label">导航计时:</span><span class="info-value" id="navigationTiming">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🎵 媒体能力</h3>
                <div class="info-item"><span class="info-label">媒体能力:</span><span class="info-value" id="mediaCapabilities">检测中...</span></div>
                <div class="info-item"><span class="info-label">视频编解码器:</span><span class="info-value" id="videoCodecs">检测中...</span></div>
                <div class="info-item"><span class="info-label">音频编解码器:</span><span class="info-value" id="audioCodecs">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🎮 WebGL详情</h3>
                <div class="info-item"><span class="info-label">WebGL供应商:</span><span class="info-value" id="webglVendor">检测中...</span></div>
                <div class="info-item"><span class="info-label">WebGL渲染器:</span><span class="info-value" id="webglRenderer">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>⌨️ 输入设备</h3>
                <div class="info-item"><span class="info-label">键盘布局:</span><span class="info-value" id="keyboardLayout">检测中...</span></div>
                <div class="info-item"><span class="info-label">指针类型:</span><span class="info-value" id="pointerType">检测中...</span></div>
                <div class="info-item"><span class="info-label">悬停能力:</span><span class="info-value" id="hoverCapable">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🌍 语言与时间</h3>
                <div class="info-item"><span class="info-label">语言列表:</span><span class="info-value" id="languages">检测中...</span></div>
                <div class="info-item"><span class="info-label">时区偏移:</span><span class="info-value" id="timezoneOffset">检测中...</span></div>
                <div class="info-item"><span class="info-label">系统时间:</span><span class="info-value" id="systemTime">检测中...</span></div>
            </div>

            <div class="info-card">
                <h3>🔊 音频指纹</h3>
                <div class="info-item"><span class="info-label">音频指纹:</span><span class="info-value" id="audioFingerprint">检测中...</span></div>
            </div>

        </div>
        
        <div class="rating-card">
            <h3>⭐ 给这个项目打个分</h3>
            <div class="stars" id="starRow">
                <span class="star" data-v="1" onclick="submitRating(1)" role="button" tabindex="0" aria-label="1星">★</span>
                <span class="star" data-v="2" onclick="submitRating(2)" role="button" tabindex="0" aria-label="2星">★</span>
                <span class="star" data-v="3" onclick="submitRating(3)" role="button" tabindex="0" aria-label="3星">★</span>
                <span class="star" data-v="4" onclick="submitRating(4)" role="button" tabindex="0" aria-label="4星">★</span>
                <span class="star" data-v="5" onclick="submitRating(5)" role="button" tabindex="0" aria-label="5星">★</span>
            </div>
            <div class="rating-msg" id="ratingMsg">点击星星进行评分</div>
            <div class="rating-stats" id="ratingStats"></div>
        </div>

        <div class="actions">
            <button class="btn" onclick="collectDeviceInfo()">🔄 重新收集信息</button>
        </div>
        
        <div style="text-align: center; color: white; margin-top: 20px; padding: 20px; font-size: 0.9rem; opacity: 0.8;">
            如您有疑问可联系daoxuans
        </div>
    </div>

    <script>
        async function collectDeviceInfo() {
            const statusElement = document.getElementById('status');
            statusElement.className = 'status';
            statusElement.textContent = '正在收集设备信息...';
            
            try {
                const deviceInfo = {
                    // 基础信息
                    userAgent: navigator.userAgent,
                    screen: screen.width + " x " + screen.height,
                    availableScreen: screen.availWidth + " x " + screen.availHeight,
                    colorDepth: screen.colorDepth + " bit",
                    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                    language: navigator.language,
                    platform: navigator.platform,
                    cpuCores: navigator.hardwareConcurrency ? navigator.hardwareConcurrency.toString() : '未知',
                    deviceMemory: navigator.deviceMemory ? navigator.deviceMemory + " GB" : '未知',
                    connection: getConnectionInfo(),
                    touchSupport: 'ontouchstart' in window ? '支持' : '不支持',
                    pixelRatio: window.devicePixelRatio.toString(),
                    cookiesEnabled: navigator.cookieEnabled ? '启用' : '禁用',
                    javaEnabled: typeof navigator.javaEnabled === 'function' ? (navigator.javaEnabled() ? '启用' : '禁用') : '未知',
                    doNotTrack: navigator.doNotTrack || '未设置',
                    hardwareConcurrency: navigator.hardwareConcurrency ? navigator.hardwareConcurrency.toString() : '未知',
                    vendor: navigator.vendor || '未知',
                    product: navigator.product || '未知',
                    
                    // 新增信息
                    battery: getBatteryInfo(),
                    onlineStatus: navigator.onLine ? '在线' : '离线',
                    maxTouchPoints: navigator.maxTouchPoints ? navigator.maxTouchPoints.toString() : '0',
                    pdfViewer: checkPDFViewer(),
                    webgl: checkWebGL(),
                    canvas: checkCanvas(),
                    audioContext: checkAudioContext(),
                    localStorage: checkLocalStorage(),
                    sessionStorage: checkSessionStorage(),
                    indexedDB: 'indexedDB' in window ? '支持' : '不支持',
                    geolocation: 'geolocation' in navigator ? '支持' : '不支持',
                    locationDetails: getLocationDetails(),
                    notifications: 'Notification' in window ? '支持' : '不支持',
                    serviceWorker: 'serviceWorker' in navigator ? '支持' : '不支持',
                    webrtc: checkWebRTC(),
                    mediaDevices: 'mediaDevices' in navigator ? '支持' : '不支持',
                    deviceOrientation: 'DeviceOrientationEvent' in window ? '支持' : '不支持',
                    vibration: 'vibrate' in navigator ? '支持' : '不支持',
                    clipboard: 'clipboard' in navigator ? '支持' : '不支持',
                    accelerometer: checkAccelerometer(),
                    gyroscope: checkGyroscope(),
                    magnetometer: 'Magnetometer' in window ? '支持' : '不支持',
                    gamepadAPI: 'getGamepads' in navigator ? '支持' : '不支持',
                    vrDisplay: 'getVRDisplays' in navigator ? '支持' : '不支持',
                    webAssembly: 'WebAssembly' in window ? '支持' : '不支持',
                    cssFeatures: getCSSFeatures(),
                    fontList: getFontList(),
                    plugins: getPluginsList(),
                    mimeTypes: getMimeTypesList(),
                    viewportSize: window.innerWidth + " x " + window.innerHeight,
                    deviceType: getDeviceType(),
                    osVersion: getOSVersion(),
                    browserVersion: getBrowserVersion(),
                    referrerPolicy: document.referrerPolicy || '未设置',
                    httpsSupport: location.protocol === 'https:' ? '支持' : '不支持',
                    // Canvas指纹
                    canvasFingerprint: generateCanvasFingerprint(),
                    webglFingerprint: generateWebGLFingerprint(),
                    fontFingerprint: generateFontFingerprint(),
                    // 音频指纹
                    audioFingerprint: getAudioFingerprint(),
                    // 屏幕详细信息
                    screenOrientation: getScreenOrientation(),
                    colorGamut: getColorGamut(),
                    hdr: getHDR(),
                    refreshRate: await getRefreshRate(),
                    // 网络详细信息
                    downlinkMax: getDownlinkMax(),
                    networkType: getNetworkType(),
                    // 性能信息
                    memoryInfo: getMemoryInfo(),
                    navigationTiming: getNavigationTiming(),
                    // 键盘布局
                    keyboardLayout: await getKeyboardLayout(),
                    // 语言偏好
                    languages: getLanguages(),
                    // 媒体能力
                    mediaCapabilities: await getMediaCapabilities(),
                    videoCodecs: await getVideoCodecs(),
                    audioCodecs: await getAudioCodecs(),
                    // WebGL详细信息
                    webglVendor: getWebGLVendor(),
                    webglRenderer: getWebGLRenderer(),
                    // 安全与隐私
                    adBlocker: await detectAdBlocker(),
                    privateMode: await detectPrivateMode(),
                    // 时间信息
                    timezoneOffset: getTimezoneOffset(),
                    systemTime: getSystemTime(),
                    // Pointer能力
                    pointerType: getPointerType(),
                    hoverCapable: getHoverCapable(),
                    // 动画帧率
                    animationFrameRate: await getAnimationFrameRate()
                };
                
                console.log('准备发送的数据:', deviceInfo);
                
                updateDisplay(deviceInfo);
                
                const controller = new AbortController();
                const timeoutId = setTimeout(() => controller.abort(), 10000);
                
                fetch('/collect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(deviceInfo),
                    signal: controller.signal
                })
                .then(response => {
                    clearTimeout(timeoutId);
                    console.log('服务器响应状态:', response.status);
                    if (!response.ok) {
                        return response.text().then(text => {
                            console.log('服务器错误响应:', text);
                            throw new Error('HTTP ' + response.status + ': ' + response.statusText);
                        });
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('服务器成功响应:', data);
                    if (data.status === 'success') {
                        statusElement.className = 'status success';
                        statusElement.textContent = '✅ 设备信息收集成功！数据已发送到服务器。';
                        if (data.data) {
                            document.getElementById('ipAddress').textContent = data.data.ipAddress || '未知';
                            document.getElementById('timestamp').textContent = data.data.timestamp || '未知';
                        }
                    } else {
                        throw new Error(data.message || '未知错误');
                    }
                })
                .catch(error => {
                    clearTimeout(timeoutId);
                    statusElement.className = 'status error';
                    console.error('请求错误:', error);
                    if (error.name === 'AbortError') {
                        statusElement.textContent = '❌ 请求超时，请检查网络连接';
                    } else {
                        statusElement.textContent = '❌ 发送数据到服务器失败: ' + error.message;
                    }
                });
            } catch (error) {
                statusElement.className = 'status error';
                statusElement.textContent = '❌ 收集设备信息时发生错误: ' + error.message;
                console.error('收集信息错误:', error);
            }
        }
        
        // 辅助函数
        function getBatteryInfo() {
            if ('getBattery' in navigator) {
                navigator.getBattery().then(function(battery) {
                    const level = Math.round(battery.level * 100);
                    const charging = battery.charging ? '充电中' : '未充电';
                    const batteryStr = level + '% (' + charging + ')';
                    
                    const element = document.getElementById('battery');
                    if (element) {
                        element.textContent = batteryStr;
                    }
                }).catch(function() {
                    const element = document.getElementById('battery');
                    if (element) {
                        element.textContent = 'API支持但获取失败';
                    }
                });
                return 'API支持，获取中...';
            }
            return '不支持';
        }
        
        function checkWebGL() {
            try {
                const canvas = document.createElement('canvas');
                return !!(window.WebGLRenderingContext && canvas.getContext('webgl')) ? '支持' : '不支持';
            } catch (e) {
                return '不支持';
            }
        }
        
        function checkCanvas() {
            try {
                const canvas = document.createElement('canvas');
                return !!(canvas.getContext && canvas.getContext('2d')) ? '支持' : '不支持';
            } catch (e) {
                return '不支持';
            }
        }
        
        function checkAudioContext() {
            return !!(window.AudioContext || window.webkitAudioContext) ? '支持' : '不支持';
        }
        
        function checkLocalStorage() {
            try {
                return 'localStorage' in window && window.localStorage !== null ? '支持' : '不支持';
            } catch (e) {
                return '不支持';
            }
        }
        
        function checkSessionStorage() {
            try {
                return 'sessionStorage' in window && window.sessionStorage !== null ? '支持' : '不支持';
            } catch (e) {
                return '不支持';
            }
        }
        
        function checkAccelerometer() {
            // 检测加速度计支持的多种方式
            const checks = [];
            
            // 方式1: 检查DeviceMotionEvent（最常用的方式）
            if (typeof DeviceMotionEvent !== 'undefined') {
                checks.push('DeviceMotion API');
            }
            
            // 方式2: 检查新的Sensor API
            if (typeof Accelerometer !== 'undefined') {
                checks.push('Accelerometer API');
            }
            
            // 方式3: 检查是否可以监听devicemotion事件
            if (window.ondevicemotion !== undefined) {
                checks.push('事件监听');
            }
            
            // 移动设备通常都有加速度计
            if (checks.length === 0 && /Mobile|Android|iPhone|iPad/i.test(navigator.userAgent)) {
                return '支持（移动设备）';
            }
            
            return checks.length > 0 ? '支持 (' + checks.join(', ') + ')' : '不支持';
        }
        
        function checkGyroscope() {
            // 检测陀螺仪支持的多种方式
            const checks = [];
            
            // 方式1: 检查DeviceOrientationEvent（最常用的方式）
            if (typeof DeviceOrientationEvent !== 'undefined') {
                checks.push('DeviceOrientation API');
            }
            
            // 方式2: 检查新的Sensor API
            if (typeof Gyroscope !== 'undefined') {
                checks.push('Gyroscope API');
            }
            
            // 方式3: 检查是否可以监听deviceorientation事件
            if (window.ondeviceorientation !== undefined) {
                checks.push('事件监听');
            }
            
            // 移动设备通常都有陀螺仪
            if (checks.length === 0 && /Mobile|Android|iPhone|iPad/i.test(navigator.userAgent)) {
                return '支持（移动设备）';
            }
            
            return checks.length > 0 ? '支持 (' + checks.join(', ') + ')' : '不支持';
        }
        
        function checkPDFViewer() {
            // 检查多种 PDF 支持方式
            const checks = [];
            
            // 检查 MIME 类型
            if (navigator.mimeTypes && navigator.mimeTypes['application/pdf']) {
                checks.push('MIME支持');
            }
            
            // 检查插件
            if (navigator.plugins) {
                for (let i = 0; i < navigator.plugins.length; i++) {
                    const plugin = navigator.plugins[i];
                    if (plugin.name.toLowerCase().includes('pdf')) {
                        checks.push('插件支持');
                        break;
                    }
                }
            }
            
            // 检查内置 PDF 查看器
            if (window.navigator.pdfViewerEnabled !== undefined) {
                if (window.navigator.pdfViewerEnabled) {
                    checks.push('内置查看器');
                }
            } else {
                // Firefox/Chrome 的内置 PDF 支持
                const userAgent = navigator.userAgent.toLowerCase();
                if (userAgent.includes('firefox') || userAgent.includes('chrome') || userAgent.includes('edge')) {
                    checks.push('可能支持内置');
                }
            }
            
            return checks.length > 0 ? checks.join(', ') : '不支持';
        }
        
        function checkWebRTC() {
            return !!(window.RTCPeerConnection || window.webkitRTCPeerConnection || window.mozRTCPeerConnection) ? '支持' : '不支持';
        }
        
        function getCSSFeatures() {
            const features = [];
            if (CSS.supports('display', 'grid')) features.push('Grid');
            if (CSS.supports('display', 'flex')) features.push('Flexbox');
            if (CSS.supports('backdrop-filter', 'blur(10px)')) features.push('Backdrop-filter');
            return features.length ? features.join(', ') : '基础CSS';
        }
        
        function getFontList() {
            const fonts = ['Arial', 'Times New Roman', 'Helvetica', 'Georgia', 'Verdana'];
            const available = [];
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            fonts.forEach(font => {
                ctx.font = '12px ' + font;
                if (ctx.measureText('test').width > 0) available.push(font);
            });
            return available.length ? available.slice(0, 3).join(', ') + '等' : '未检测';
        }
        
        function getPluginsList() {
            if (navigator.plugins && navigator.plugins.length > 0) {
                const plugins = Array.from(navigator.plugins).slice(0, 3).map(p => p.name);
                return plugins.join(', ') + '等';
            }
            return '无插件';
        }
        
        function getMimeTypesList() {
            if (navigator.mimeTypes && navigator.mimeTypes.length > 0) {
                return navigator.mimeTypes.length + ' 种类型';
            }
            return '未检测';
        }
        
        function getDeviceType() {
            const ua = navigator.userAgent.toLowerCase();
            if (/mobile|android|iphone|ipad|phone/i.test(ua)) return '移动设备';
            if (/tablet|ipad/i.test(ua)) return '平板设备';
            return '桌面设备';
        }
        
        function getOSVersion() {
            const ua = navigator.userAgent;
            if (ua.indexOf('Windows NT 10.0') !== -1) return 'Windows 10/11';
            if (ua.indexOf('Windows NT 6.3') !== -1) return 'Windows 8.1';
            if (ua.indexOf('Windows NT 6.2') !== -1) return 'Windows 8';
            if (ua.indexOf('Windows NT 6.1') !== -1) return 'Windows 7';
            if (ua.indexOf('Mac OS X') !== -1) return 'macOS ' + ua.match(/Mac OS X ([0-9_]+)/)?.[1]?.replace(/_/g, '.') || '';
            if (ua.indexOf('Android') !== -1) return 'Android ' + ua.match(/Android ([0-9.]+)/)?.[1] || '';
            if (ua.indexOf('iPhone OS') !== -1) return 'iOS ' + ua.match(/iPhone OS ([0-9_]+)/)?.[1]?.replace(/_/g, '.') || '';
            return navigator.platform;
        }
        
        function getBrowserVersion() {
            const ua = navigator.userAgent;
            if (ua.indexOf('Chrome') !== -1) return 'Chrome ' + ua.match(/Chrome\/([0-9.]+)/)?.[1] || '';
            if (ua.indexOf('Firefox') !== -1) return 'Firefox ' + ua.match(/Firefox\/([0-9.]+)/)?.[1] || '';
            if (ua.indexOf('Safari') !== -1 && ua.indexOf('Chrome') === -1) return 'Safari ' + ua.match(/Version\/([0-9.]+)/)?.[1] || '';
            if (ua.indexOf('Edge') !== -1) return 'Edge ' + ua.match(/Edge\/([0-9.]+)/)?.[1] || '';
            return '未知浏览器';
        }
        
        function getConnectionInfo() {
            // 检查Network Information API支持
            if (navigator.connection || navigator.mozConnection || navigator.webkitConnection) {
                const conn = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
                
                let connectionType = '未知';
                let speedInfo = '';
                
                // 只使用type字段判断连接类型（effectiveType仅表示速度等级，不表示连接类型）
                if (conn.type) {
                    switch(conn.type) {
                        case 'cellular':
                            connectionType = '蜂窝网络';
                            break;
                        case 'wifi':
                            connectionType = 'WiFi';
                            break;
                        case 'ethernet':
                            connectionType = '有线网络';
                            break;
                        case 'bluetooth':
                            connectionType = '蓝牙';
                            break;
                        case 'wimax':
                            connectionType = 'WiMAX';
                            break;
                        case 'none':
                            connectionType = '无连接';
                            break;
                        default:
                            connectionType = '其他';
                    }
                } else {
                    // 如果没有type字段，显示effectiveType作为参考
                    if (conn.effectiveType) {
                        connectionType = '未知 (网速等级: ' + conn.effectiveType + ')';
                    }
                }
                
                // 添加速度信息
                const speedParts = [];
                if (conn.downlink !== undefined && conn.downlink > 0) {
                    speedParts.push('下行' + conn.downlink + 'Mbps');
                }
                if (conn.rtt !== undefined && conn.rtt > 0) {
                    speedParts.push('延迟' + conn.rtt + 'ms');
                }
                
                if (speedParts.length > 0) {
                    speedInfo = ' (' + speedParts.join(', ') + ')';
                }
                
                return connectionType + speedInfo;
            }
            
            return '不支持 Network Information API';
        }
        
        // 获取地理位置详情
        function getLocationDetails() {
            if ('geolocation' in navigator) {
                navigator.geolocation.getCurrentPosition(
                    function(position) {
                        const lat = position.coords.latitude.toFixed(6);
                        const lng = position.coords.longitude.toFixed(6);
                        const accuracy = position.coords.accuracy.toFixed(0);
                        const locationStr = '纬度: ' + lat + ', 经度: ' + lng + ' (精度: ' + accuracy + 'm)';
                        
                        // 更新显示
                        const element = document.getElementById('locationDetails');
                        if (element) {
                            element.textContent = locationStr;
                        }
                        
                        // 尝试获取地址信息（可选）
                        reverseGeocode(lat, lng);
                    },
                    function(error) {
                        const element = document.getElementById('locationDetails');
                        if (element) {
                            switch(error.code) {
                                case error.PERMISSION_DENIED:
                                    element.textContent = '用户拒绝了地理定位请求';
                                    break;
                                case error.POSITION_UNAVAILABLE:
                                    element.textContent = '位置信息不可用';
                                    break;
                                case error.TIMEOUT:
                                    element.textContent = '请求用户地理位置超时';
                                    break;
                                default:
                                    element.textContent = '发生未知错误';
                                    break;
                            }
                        }
                    },
                    {
                        enableHighAccuracy: true,
                        timeout: 10000,
                        maximumAge: 60000
                    }
                );
                return '正在获取位置...';
            }
            return '不支持地理位置API';
        }
        
        // 反向地理编码（可选功能）
        function reverseGeocode(lat, lng) {
            // 注意：这里使用免费的API，实际使用时可能需要API密钥
            fetch('https://nominatim.openstreetmap.org/reverse?format=json&lat=' + lat + '&lon=' + lng + '&zoom=18&addressdetails=1')
                .then(response => response.json())
                .then(data => {
                    if (data && data.display_name) {
                        const element = document.getElementById('locationDetails');
                        if (element) {
                            const currentText = element.textContent;
                            element.textContent = currentText + ' - ' + data.display_name;
                        }
                    }
                })
                .catch(error => {
                    console.log('反向地理编码失败:', error);
                });
        }
        
        // Canvas指纹生成函数
        function generateCanvasFingerprint() {
            try {
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');
                
                if (!ctx) return '不支持';
                
                // 设置Canvas尺寸
                canvas.width = 300;
                canvas.height = 150;
                
                // 绘制背景渐变
                const gradient = ctx.createLinearGradient(0, 0, canvas.width, canvas.height);
                gradient.addColorStop(0, '#ff6b6b');
                gradient.addColorStop(0.5, '#4ecdc4');
                gradient.addColorStop(1, '#45b7d1');
                ctx.fillStyle = gradient;
                ctx.fillRect(0, 0, canvas.width, canvas.height);
                
                // 绘制几何形状
                ctx.strokeStyle = '#333333';
                ctx.lineWidth = 2;
                ctx.strokeRect(10, 10, 100, 50);
                
                ctx.fillStyle = '#ff9999';
                ctx.beginPath();
                ctx.arc(180, 80, 40, 0, Math.PI * 2);
                ctx.fill();
                
                // 绘制文本 - 使用不同字体和样式
                ctx.fillStyle = '#333333';
                ctx.font = '16px Arial';
                ctx.fillText('Device Fingerprint', 10, 80);
                
                ctx.font = 'bold 12px serif';
                ctx.fillText('Canvas Test 2024', 10, 100);
                
                ctx.font = '14px monospace';
                ctx.fillText('Hello World! 你好世界', 10, 120);
                
                // 绘制表情符号
                ctx.font = '20px Arial';
                ctx.fillText('😀🌍🔒', 200, 120);
                
                // 添加阴影效果
                ctx.shadowColor = 'rgba(0,0,0,0.5)';
                ctx.shadowBlur = 5;
                ctx.shadowOffsetX = 3;
                ctx.shadowOffsetY = 3;
                ctx.fillStyle = '#4a90e2';
                ctx.fillRect(220, 20, 60, 30);
                
                // 生成Canvas数据URL并计算哈希
                const dataURL = canvas.toDataURL();
                return hashString(dataURL);
            } catch (e) {
                return '生成失败: ' + e.message;
            }
        }

        // WebGL指纹生成函数
        function generateWebGLFingerprint() {
            try {
                const canvas = document.createElement('canvas');
                const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
                
                if (!gl) return '不支持';
                
                const fingerprint = [];
                
                // WebGL版本和供应商信息
                fingerprint.push(gl.getParameter(gl.VERSION));
                fingerprint.push(gl.getParameter(gl.VENDOR));
                fingerprint.push(gl.getParameter(gl.RENDERER));
                fingerprint.push(gl.getParameter(gl.SHADING_LANGUAGE_VERSION));
                
                // 扩展信息
                const extensions = gl.getSupportedExtensions();
                if (extensions) {
                    fingerprint.push(extensions.sort().join(','));
                }
                
                // WebGL参数
                const params = [
                    gl.MAX_TEXTURE_SIZE,
                    gl.MAX_VERTEX_ATTRIBS,
                    gl.MAX_VERTEX_UNIFORM_VECTORS,
                    gl.MAX_FRAGMENT_UNIFORM_VECTORS,
                    gl.MAX_VARYING_VECTORS,
                    gl.MAX_RENDERBUFFER_SIZE,
                    gl.MAX_VIEWPORT_DIMS
                ];
                
                params.forEach(param => {
                    fingerprint.push(gl.getParameter(param));
                });
                
                // 生成简单的WebGL渲染
                gl.clearColor(0.2, 0.4, 0.8, 1.0);
                gl.clear(gl.COLOR_BUFFER_BIT);
                
                return hashString(fingerprint.join('|'));
            } catch (e) {
                return '生成失败: ' + e.message;
            }
        }

        // 字体指纹生成函数
        function generateFontFingerprint() {
            try {
                const baseFonts = ['monospace', 'sans-serif', 'serif'];
                const testFonts = [
                    'Arial', 'Helvetica', 'Times New Roman', 'Courier New', 'Verdana',
                    'Georgia', 'Palatino', 'Garamond', 'Bookman', 'Comic Sans MS',
                    'Trebuchet MS', 'Arial Black', 'Impact', 'Tahoma', 'Geneva',
                    'Lucida Console', 'Monaco', 'Consolas', 'Calibri', 'Cambria',
                    'Microsoft YaHei', 'SimSun', 'SimHei', 'KaiTi', 'FangSong'
                ];
                
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');
                
                if (!ctx) return '不支持';
                
                const detectedFonts = [];
                const testString = 'mmmmmmmmmmlli';
                const testSize = '72px';
                
                // 为每种基础字体创建基准尺寸
                const baseSizes = {};
                baseFonts.forEach(baseFont => {
                    ctx.font = testSize + ' ' + baseFont;
                    const metrics = ctx.measureText(testString);
                    baseSizes[baseFont] = {
                        width: metrics.width,
                        height: metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent
                    };
                });
                
                // 测试每种字体
                testFonts.forEach(font => {
                    baseFonts.forEach(baseFont => {
                        ctx.font = testSize + ' ' + font + ', ' + baseFont;
                        const metrics = ctx.measureText(testString);
                        const currentSize = {
                            width: metrics.width,
                            height: metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent
                        };
                        
                        // 如果尺寸与基础字体不同，说明目标字体存在
                        if (currentSize.width !== baseSizes[baseFont].width || 
                            currentSize.height !== baseSizes[baseFont].height) {
                            if (!detectedFonts.includes(font)) {
                                detectedFonts.push(font);
                            }
                        }
                    });
                });
                
                return hashString(detectedFonts.sort().join(','));
            } catch (e) {
                return '生成失败: ' + e.message;
            }
        }

        // 简单的哈希函数
        function hashString(str) {
            let hash = 0;
            if (str.length === 0) return hash.toString();
            
            for (let i = 0; i < str.length; i++) {
                const char = str.charCodeAt(i);
                hash = ((hash << 5) - hash) + char;
                hash = hash & hash; // 转换为32位整数
            }
            
            return Math.abs(hash).toString(16);
        }

        // ========== 新增检测函数 ==========
        
        // 音频指纹
        function getAudioFingerprint() {
            try {
                const AudioContext = window.AudioContext || window.webkitAudioContext;
                if (!AudioContext) return '不支持';
                
                const context = new AudioContext();
                const oscillator = context.createOscillator();
                const analyser = context.createAnalyser();
                const gainNode = context.createGain();
                const scriptProcessor = context.createScriptProcessor(4096, 1, 1);
                
                gainNode.gain.value = 0; // 静音
                oscillator.type = 'triangle';
                oscillator.connect(analyser);
                analyser.connect(scriptProcessor);
                scriptProcessor.connect(gainNode);
                gainNode.connect(context.destination);
                
                oscillator.start(0);
                const fingerprint = [];
                
                scriptProcessor.onaudioprocess = function(event) {
                    const output = event.outputBuffer.getChannelData(0);
                    for (let i = 0; i < output.length; i++) {
                        fingerprint.push(output[i].toString());
                    }
                    
                    if (fingerprint.length > 30) {
                        oscillator.disconnect();
                        scriptProcessor.disconnect();
                    }
                };
                
                setTimeout(() => {
                    context.close();
                }, 100);
                
                return hashString(fingerprint.slice(0, 30).join(','));
            } catch (e) {
                return '生成失败';
            }
        }

        // 屏幕方向
        function getScreenOrientation() {
            if (screen.orientation) {
                return screen.orientation.type || '未知';
            }
            if (window.orientation !== undefined) {
                return window.orientation === 0 ? 'portrait' : 'landscape';
            }
            return '不支持';
        }

        // 色域
        function getColorGamut() {
            if (window.matchMedia) {
                if (window.matchMedia('(color-gamut: rec2020)').matches) return 'rec2020';
                if (window.matchMedia('(color-gamut: p3)').matches) return 'p3';
                if (window.matchMedia('(color-gamut: srgb)').matches) return 'srgb';
            }
            return '未知';
        }

        // HDR支持
        function getHDR() {
            if (window.matchMedia) {
                if (window.matchMedia('(dynamic-range: high)').matches) return '支持';
            }
            return '不支持';
        }

        // 刷新率
        function getRefreshRate() {
            return new Promise((resolve) => {
                let lastTime = performance.now();
                let frames = 0;
                
                function measureFrame() {
                    const currentTime = performance.now();
                    frames++;
                    
                    if (frames >= 60 || currentTime - lastTime >= 1000) {
                        const fps = Math.round((frames * 1000) / (currentTime - lastTime));
                        resolve(fps + ' Hz');
                    } else {
                        requestAnimationFrame(measureFrame);
                    }
                }
                
                requestAnimationFrame(measureFrame);
                setTimeout(() => resolve('测量超时'), 2000);
            });
        }

        // 最大下行速度
        function getDownlinkMax() {
            const conn = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
            if (conn && conn.downlinkMax) {
                return conn.downlinkMax + ' Mbps';
            }
            return '未知';
        }

        // 网络类型详细
        function getNetworkType() {
            const conn = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
            if (conn) {
                const parts = [];
                // type是真实的连接类型（wifi, ethernet, cellular等）
                if (conn.type) {
                    parts.push('连接类型: ' + conn.type);
                }
                // effectiveType是网络速度等级（slow-2g, 2g, 3g, 4g）
                if (conn.effectiveType) {
                    parts.push('速度等级: ' + conn.effectiveType);
                }
                // saveData表示用户是否启用了数据节省模式
                if (conn.saveData !== undefined) {
                    parts.push('省流模式: ' + (conn.saveData ? '开启' : '关闭'));
                }
                return parts.length > 0 ? parts.join(', ') : '未知';
            }
            return '不支持';
        }

        // 内存信息
        function getMemoryInfo() {
            if (performance.memory) {
                const used = (performance.memory.usedJSHeapSize / 1048576).toFixed(2);
                const total = (performance.memory.totalJSHeapSize / 1048576).toFixed(2);
                const limit = (performance.memory.jsHeapSizeLimit / 1048576).toFixed(2);
                return '已用: ' + used + 'MB, 总计: ' + total + 'MB, 限制: ' + limit + 'MB';
            }
            return '不支持';
        }

        // 导航计时
        function getNavigationTiming() {
            if (performance.timing) {
                const timing = performance.timing;
                const loadTime = timing.loadEventEnd - timing.navigationStart;
                const domReady = timing.domContentLoadedEventEnd - timing.navigationStart;
                return '加载时间: ' + loadTime + 'ms, DOM准备: ' + domReady + 'ms';
            }
            return '不支持';
        }

        // 键盘布局
        function getKeyboardLayout() {
            return new Promise((resolve) => {
                if (navigator.keyboard && navigator.keyboard.getLayoutMap) {
                    navigator.keyboard.getLayoutMap()
                        .then(layoutMap => {
                            const entries = Array.from(layoutMap.entries()).slice(0, 5);
                            resolve(entries.length > 0 ? '支持 (' + entries.length + '个键位)' : '支持');
                        })
                        .catch(() => resolve('获取失败'));
                } else {
                    resolve('不支持');
                }
            });
        }

        // 语言列表
        function getLanguages() {
            if (navigator.languages) {
                return navigator.languages.join(', ');
            }
            return navigator.language || '未知';
        }

        // 媒体能力
        function getMediaCapabilities() {
            return new Promise((resolve) => {
                if (navigator.mediaCapabilities) {
                    const config = {
                        type: 'file',
                        video: {
                            contentType: 'video/mp4; codecs="avc1.42E01E"',
                            width: 1920,
                            height: 1080,
                            bitrate: 2000000,
                            framerate: 30
                        }
                    };
                    
                    navigator.mediaCapabilities.decodingInfo(config)
                        .then(result => {
                            resolve('支持 (平滑: ' + result.smooth + ', 省电: ' + result.powerEfficient + ')');
                        })
                        .catch(() => resolve('检测失败'));
                } else {
                    resolve('不支持');
                }
            });
        }

        // 视频编解码器
        function getVideoCodecs() {
            return new Promise((resolve) => {
                const codecs = ['avc1.42E01E', 'vp8', 'vp9', 'av01.0.05M.08', 'hev1.1.6.L93.B0'];
                const supported = [];
                
                codecs.forEach(codec => {
                    const canPlay = document.createElement('video').canPlayType('video/mp4; codecs="' + codec + '"');
                    if (canPlay) supported.push(codec);
                });
                
                resolve(supported.length > 0 ? supported.join(', ') : '无支持编解码器');
            });
        }

        // 音频编解码器
        function getAudioCodecs() {
            return new Promise((resolve) => {
                const codecs = ['opus', 'vorbis', 'mp4a.40.2', 'flac', 'mp3'];
                const supported = [];
                
                codecs.forEach(codec => {
                    const canPlay = document.createElement('audio').canPlayType('audio/' + (codec === 'mp4a.40.2' ? 'mp4' : codec === 'mp3' ? 'mpeg' : codec) + (codec === 'mp4a.40.2' ? '; codecs="' + codec + '"' : ''));
                    if (canPlay) supported.push(codec);
                });
                
                resolve(supported.length > 0 ? supported.join(', ') : '无支持编解码器');
            });
        }

        // WebGL供应商
        function getWebGLVendor() {
            try {
                const canvas = document.createElement('canvas');
                const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
                if (gl) {
                    const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
                    if (debugInfo) {
                        return gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL);
                    }
                }
                return '未知';
            } catch (e) {
                return '获取失败';
            }
        }

        // WebGL渲染器
        function getWebGLRenderer() {
            try {
                const canvas = document.createElement('canvas');
                const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
                if (gl) {
                    const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
                    if (debugInfo) {
                        return gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL);
                    }
                }
                return '未知';
            } catch (e) {
                return '获取失败';
            }
        }

        // 广告拦截检测
        function detectAdBlocker() {
            return new Promise((resolve) => {
                const testAd = document.createElement('div');
                testAd.innerHTML = '&nbsp;';
                testAd.className = 'adsbox ad-placement carbon-ads';
                testAd.style.position = 'absolute';
                testAd.style.left = '-999px';
                document.body.appendChild(testAd);
                
                setTimeout(() => {
                    const detected = testAd.offsetHeight === 0 || window.getComputedStyle(testAd).display === 'none';
                    document.body.removeChild(testAd);
                    resolve(detected ? '检测到' : '未检测到');
                }, 100);
            });
        }

        // 隐私模式检测
        function detectPrivateMode() {
            return new Promise((resolve) => {
                if (navigator.storage && navigator.storage.estimate) {
                    navigator.storage.estimate().then(estimate => {
                        // 隐私模式下配额通常很小
                        resolve(estimate.quota < 120000000 ? '可能是' : '否');
                    }).catch(() => resolve('检测失败'));
                } else {
                    resolve('不支持');
                }
            });
        }

        // 时区偏移
        function getTimezoneOffset() {
            const offset = new Date().getTimezoneOffset();
            const hours = Math.floor(Math.abs(offset) / 60);
            const minutes = Math.abs(offset) % 60;
            const sign = offset > 0 ? '-' : '+';
            return 'UTC' + sign + hours.toString().padStart(2, '0') + ':' + minutes.toString().padStart(2, '0');
        }

        // 系统时间
        function getSystemTime() {
            return new Date().toLocaleString('zh-CN', { 
                year: 'numeric', 
                month: '2-digit', 
                day: '2-digit', 
                hour: '2-digit', 
                minute: '2-digit', 
                second: '2-digit' 
            });
        }

        // 指针类型
        function getPointerType() {
            if (window.matchMedia) {
                if (window.matchMedia('(pointer: fine)').matches) return 'fine (鼠标)';
                if (window.matchMedia('(pointer: coarse)').matches) return 'coarse (触摸)';
                if (window.matchMedia('(pointer: none)').matches) return 'none';
            }
            return '未知';
        }

        // 悬停能力
        function getHoverCapable() {
            if (window.matchMedia) {
                if (window.matchMedia('(hover: hover)').matches) return '支持';
                if (window.matchMedia('(hover: none)').matches) return '不支持';
            }
            return '未知';
        }

        // 动画帧率
        function getAnimationFrameRate() {
            return new Promise((resolve) => {
                let lastTime = performance.now();
                let frameCount = 0;
                const frameTimes = [];
                
                function measureFrame(currentTime) {
                    frameTimes.push(currentTime - lastTime);
                    lastTime = currentTime;
                    frameCount++;
                    
                    if (frameCount >= 30) {
                        const avgFrameTime = frameTimes.reduce((a, b) => a + b) / frameTimes.length;
                        const fps = Math.round(1000 / avgFrameTime);
                        resolve(fps + ' FPS');
                    } else {
                        requestAnimationFrame(measureFrame);
                    }
                }
                
                requestAnimationFrame(measureFrame);
                setTimeout(() => resolve('测量超时'), 2000);
            });
        }
        
        function updateDisplay(info) {
            for (const [key, value] of Object.entries(info)) {
                const element = document.getElementById(key);
                if (element) element.textContent = value;
            }
        }
        
        // 评分功能
        function highlightStars(n) {
            document.querySelectorAll('.star').forEach(s => {
                s.classList.toggle('active', parseInt(s.dataset.v) <= n);
            });
        }
        document.querySelectorAll('.star').forEach(s => {
            s.addEventListener('mouseover', () => highlightStars(parseInt(s.dataset.v)));
            s.addEventListener('mouseleave', () => {
                const current = parseInt(document.getElementById('starRow').dataset.selected || 0);
                highlightStars(current);
            });
            s.addEventListener('keydown', e => {
                if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); submitRating(parseInt(s.dataset.v)); }
            });
        });

        async function submitRating(score) {
            document.getElementById('starRow').dataset.selected = score;
            highlightStars(score);
            const msg = document.getElementById('ratingMsg');
            msg.textContent = '提交中...';
            try {
                const res = await fetch('/rate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ score: score })
                });
                const data = await res.json();
                msg.textContent = data.status === 'success' ? '✅ 感谢您的评分！' : ('❌ ' + data.message);
                loadRatingStats();
            } catch (e) {
                msg.textContent = '❌ 提交失败: ' + e.message;
            }
        }

        async function loadRatingStats() {
            try {
                const res = await fetch('/rating-stats');
                const data = await res.json();
                const container = document.getElementById('ratingStats');
                container.textContent = '';
                if (data.status === 'success' && data.data && data.data.total > 0) {
                    const s = data.data;

                    const avgDiv = document.createElement('div');
                    avgDiv.className = 'avg';
                    avgDiv.textContent = s.average.toFixed(1) + ' / 5';
                    container.appendChild(avgDiv);

                    const totalDiv = document.createElement('div');
                    totalDiv.textContent = '共 ' + s.total + ' 人评分';
                    container.appendChild(totalDiv);

                    const barsDiv = document.createElement('div');
                    barsDiv.style.cssText = 'max-width:280px;margin:10px auto';
                    for (let i = 5; i >= 1; i--) {
                        const pct = s.total > 0 ? Math.round(s.counts[i-1] / s.total * 100) : 0;
                        const row = document.createElement('div');
                        row.className = 'rating-bar-row';

                        const label = document.createElement('span');
                        label.textContent = i + '★';
                        row.appendChild(label);

                        const bg = document.createElement('div');
                        bg.className = 'rating-bar-bg';
                        const fill = document.createElement('div');
                        fill.className = 'rating-bar-fill';
                        fill.style.width = pct + '%';
                        bg.appendChild(fill);
                        row.appendChild(bg);

                        const count = document.createElement('span');
                        count.textContent = s.counts[i-1];
                        row.appendChild(count);

                        barsDiv.appendChild(row);
                    }
                    container.appendChild(barsDiv);
                }
            } catch (e) { /* 忽略统计加载失败 */ }
        }

        document.addEventListener('DOMContentLoaded', () => { collectDeviceInfo(); loadRatingStats(); });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// 处理评分提交
func rateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		sendJSONResponse(w, http.StatusMethodNotAllowed, Response{Status: "error", Message: "Only POST method is allowed"})
		return
	}

	ip := getClientIP(r)
	if !rateLimiter.Allow(ip) {
		sendJSONResponse(w, http.StatusTooManyRequests, Response{Status: "error", Message: "请求过于频繁，请稍后再试"})
		return
	}

	var rating Rating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, Response{Status: "error", Message: "Invalid JSON format"})
		return
	}
	if rating.Score < 1 || rating.Score > 5 {
		sendJSONResponse(w, http.StatusBadRequest, Response{Status: "error", Message: "评分必须在 1 到 5 之间"})
		return
	}

	ratingsMutex.Lock()
	ratings = append(ratings, rating.Score)
	ratingsMutex.Unlock()

	fmt.Printf("收到评分 [%s] IP: %s, 分数: %d\n", time.Now().Format("2006-01-02 15:04:05"), ip, rating.Score)
	sendJSONResponse(w, http.StatusOK, Response{Status: "success", Message: "感谢您的评分！"})
}

// 获取评分统计
func ratingStatsHandler(w http.ResponseWriter, r *http.Request) {
	ratingsMutex.Lock()
	defer ratingsMutex.Unlock()

	var stats RatingStats
	sum := 0
	for _, s := range ratings {
		stats.Total++
		stats.Counts[s-1]++
		sum += s
	}
	if stats.Total > 0 {
		stats.Average = float64(sum) / float64(stats.Total)
	}

	sendJSONResponse(w, http.StatusOK, Response{Status: "success", Data: stats})
}

func main() {
	// 设置路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/collect", collectHandler)
	http.HandleFunc("/rate", rateHandler)
	http.HandleFunc("/rating-stats", ratingStatsHandler)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动信息
	fmt.Printf("🚀 设备信息收集服务器启动成功!\n")
	fmt.Printf("📊 访问地址: http://localhost:%s\n", port)
	fmt.Printf("💻 操作系统: %s\n", runtime.GOOS)
	fmt.Printf("🕒 启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("----------------------------------------\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
