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
}

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
                <h3>�🌐 浏览器信息</h3>
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
            </div>

        </div>
        
        <div class="actions">
            <button class="btn" onclick="collectDeviceInfo()">🔄 重新收集信息</button>
        </div>
    </div>

    <script>
        function collectDeviceInfo() {
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
                    accelerometer: 'Accelerometer' in window ? '支持' : '不支持',
                    gyroscope: 'Gyroscope' in window ? '支持' : '不支持',
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
                    fontFingerprint: generateFontFingerprint()
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
                
                let info = [];
                
                // 有效连接类型
                if (conn.effectiveType) {
                    const typeMap = {
                        'slow-2g': '慢速2G',
                        '2g': '2G',
                        '3g': '3G', 
                        '4g': '4G'
                    };
                    info.push(typeMap[conn.effectiveType] || conn.effectiveType);
                }
                
                // 连接类型
                if (conn.type) {
                    const connectionTypeMap = {
                        'bluetooth': '蓝牙',
                        'cellular': '蜂窝网络',
                        'ethernet': '以太网',
                        'none': '无连接',
                        'wifi': 'WiFi',
                        'wimax': 'WiMAX',
                        'other': '其他',
                        'unknown': '未知'
                    };
                    info.push(connectionTypeMap[conn.type] || conn.type);
                }
                
                // 下行速度
                if (conn.downlink !== undefined) {
                    info.push('下行: ' + conn.downlink + 'Mbps');
                }
                
                // RTT延迟
                if (conn.rtt !== undefined) {
                    info.push('RTT: ' + conn.rtt + 'ms');
                }
                
                // 节省数据模式
                if (conn.saveData !== undefined) {
                    info.push('节省数据: ' + (conn.saveData ? '开启' : '关闭'));
                }
                
                return info.length > 0 ? info.join(' | ') : '基本连接信息';
            }
            
            // 备用检测方法
            const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
            if (!connection) {
                // 通过其他方式推断连接类型
                let fallbackInfo = [];
                
                // 检查在线状态
                fallbackInfo.push(navigator.onLine ? '在线' : '离线');
                
                // 检查是否可能是移动设备
                if (/Mobile|Android|iPhone|iPad/i.test(navigator.userAgent)) {
                    fallbackInfo.push('可能是移动网络');
                } else {
                    fallbackInfo.push('可能是宽带');
                }
                
                return fallbackInfo.join(' | ');
            }
            
            return '无法检测连接信息';
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
        
        function updateDisplay(info) {
            for (const [key, value] of Object.entries(info)) {
                const element = document.getElementById(key);
                if (element) element.textContent = value;
            }
        }
        
        document.addEventListener('DOMContentLoaded', collectDeviceInfo);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func main() {
	// 设置路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/collect", collectHandler)

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
