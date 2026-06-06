# Slot Game 集成指南

## 技术栈选择

### 后端服务（游戏逻辑/API）

| 方面 | Golang | Python |
|------|--------|--------|
| **并发性能** | ⭐⭐⭐⭐⭐ Goroutines | ⭐⭐ GIL 限制 |
| **内存占用** | ⭐⭐⭐⭐⭐ 低 | ⭐⭐⭐ 较高 |
| **启动速度** | ⭐⭐⭐⭐⭐ 秒级 | ⭐⭐⭐⭐ 秒级 |
| **开发速度** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ 快 |
| **生态成熟度** | ⭐⭐⭐⭐ 游戏服务 | ⭐⭐⭐⭐⭐ Web 框架多 |

**推荐**：游戏后端用 **Golang**
- 高并发处理 spin 请求
- 低延迟响应
- 部署简单（单二进制文件）

### 前端游戏（Godot Web Export）

**与后端无关**，导出后是纯静态文件：
```
index.html + game.js + game.wasm
```

可以放在任何静态服务器上。

---

## 集成方式

### 方案 1: iframe 嵌入（最简单）

```html
<iframe
  src="https://your-cdn.com/slot-game/index.html"
  width="800"
  height="600"
  frameborder="0"
  allowfullscreen>
</iframe>
```

**优点**：零开发成本，即插即用
**缺点**：通信受限，样式定制困难

---

### 方案 2: JavaScript SDK（推荐）

#### SDK 封装

```javascript
// slot-game-sdk.js
class SlotGameSDK {
  constructor(containerId, options = {}) {
    this.container = document.getElementById(containerId);
    this.options = options;
    this.gameFrame = null;
  }

  // 加载游戏
  load() {
    const iframe = document.createElement('iframe');
    iframe.src = this.options.gameUrl || 'https://cdn.example.com/game/index.html';
    iframe.style.width = this.options.width || '100%';
    iframe.style.height = this.options.height || '600px';
    iframe.style.border = 'none';

    this.container.appendChild(iframe);
    this.gameFrame = iframe;

    // 监听游戏消息
    window.addEventListener('message', this.handleMessage.bind(this));
  }

  // 发送消息到游戏
  sendMessage(action, data) {
    this.gameFrame.contentWindow.postMessage({
      action: action,
      data: data
    }, '*');
  }

  // 设置用户信息
  setUserInfo(userId, token) {
    this.sendMessage('setUser', { userId, token });
  }

  // 设置下注金额
  setBetAmount(amount) {
    this.sendMessage('setBet', { amount });
  }

  // 处理游戏消息
  handleMessage(event) {
    const { action, data } = event.data;

    switch(action) {
      case 'spinStart':
        this.options.onSpinStart?.(data);
        break;
      case 'spinComplete':
        this.options.onSpinComplete?.(data);
        break;
      case 'gameReady':
        this.options.onReady?.();
        break;
    }
  }

  // 销毁游戏
  destroy() {
    if (this.gameFrame) {
      this.gameFrame.remove();
    }
    window.removeEventListener('message', this.handleMessage);
  }
}

// 集成商使用方式
window.SlotGameSDK = SlotGameSDK;
```

#### 集成商使用示例

```html
<div id="slot-game"></div>

<script src="https://cdn.example.com/sdk/slot-game-sdk.js"></script>
<script>
  const game = new SlotGameSDK('slot-game', {
    gameUrl: 'https://your-cdn.com/game/',
    width: '100%',
    height: '500px',

    // 回调函数
    onReady: () => console.log('游戏已加载'),
    onSpinStart: (data) => console.log('开始旋转', data),
    onSpinComplete: (result) => {
      console.log('旋转结果', result);
      // 上报给集成商后端
      reportResult(result);
    }
  });

  // 加载游戏
  game.load();

  // 设置用户
  game.setUserInfo('user_123', 'token_xxx');
</script>
```

---

### 方案 3: Web Component（现代化）

```javascript
// slot-game-element.js
class SlotGameElement extends HTMLElement {
  connectedCallback() {
    this.render();
  }

  render() {
    const userId = this.getAttribute('user-id') || 'guest';
    const apiUrl = this.getAttribute('api-url');

    this.innerHTML = `
      <iframe src="https://cdn.example.com/game/?userId=${userId}&apiUrl=${apiUrl}"
              style="width:100%;height:600px;border:none;">
      </iframe>
    `;
  }
}

customElements.define('slot-game', SlotGameElement);
```

#### 集成商使用

```html
<slot-game user-id="user123" api-url="https://api.example.com"></slot-game>
```

---

## 推荐架构

```
┌─────────────────────────────────────────────────────┐
│                    集成商前端                          │
│  ┌─────────────────────────────────────────────┐   │
│  │            SlotGameSDK                       │   │
│  │   ┌─────────────────────────────────────┐   │   │
│  │   │         iframe (Godot Game)         │   │   │
│  │   │   ┌─────────┬─────────┬─────────┐   │   │   │
│  │   │   │ Reel 1  │ Reel 2  │ Reel 3  │   │   │   │
│  │   │   └─────────┴─────────┴─────────┘   │   │   │
│  │   └─────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
                    ↕ postMessage
┌─────────────────────────────────────────────────────┐
│              集成商后端 (Golang/Node/PHP)              │
│         ↕ (API 调用，带用户 token)                     │
┌─────────────────────────────────────────────────────┐
│                   你的游戏后端 (Golang)                │
│              /api/game/spin                           │
│              /api/game/config                         │
└─────────────────────────────────────────────────────┘
```

---

## 项目结构建议

```
your-project/
├── backend/              # Golang 游戏后端
│   ├── main.go
│   ├── api/
│   │   ├── spin.go
│   │   └── config.go
│   └── go.mod
├── game/                 # Godot 项目
│   ├── exports/          # 导出的 Web 文件
│   │   └── web/
│   │       ├── index.html
│   │       ├── game.js
│   │       └── game.wasm
│   └── scripts/
└── sdk/                  # JavaScript SDK
    ├── slot-game-sdk.js
    └── README.md
```

---

## API 接口说明

### 获取游戏配置

```http
GET /api/game/config
```

**响应**:
```json
{
  "success": true,
  "data": {
    "game_id": "game_001",
    "game_name": "Classic Slots",
    "version": "1.0.0",
    "symbols_count": 8,
    "reels_count": 3,
    "pay_lines": 5,
    "min_bet": 0.5,
    "max_bet": 100.0,
    "rtp": 0.95
  }
}
```

### Spin 请求

```http
POST /api/game/spin
Content-Type: application/json

{
  "user_id": "user_123",
  "bet_amount": 1.0,
  "bet_lines": 5,
  "session_id": "session_xxx",
  "is_free_spin": false
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "bet_amount": 1.0,
    "bet_lines": 5.0,
    "bet_per_line": 0.2,
    "win_amount": 3.0,
    "net_result": 2.0,
    "rtp_rate": 300.0,
    "reel_result": [
      ["lemon", "orange", "seven"],
      ["grape", "lemon", "watermelon"],
      ["lemon", "cherry", "lemon"]
    ],
    "win_lines": [
      {
        "line_id": 4.0,
        "symbol_id": "lemon",
        "match_count": 3.0,
        "multiplier": 15.0,
        "win_amount": 3.0,
        "positions": [0.0, 4.0, 8.0]
      }
    ]
  }
}
```

---

## 安全建议

1. **Token 验证**：所有 API 请求需携带有效 token
2. **签名验证**：关键请求使用签名防篡改
3. **限流保护**：防止暴力请求
4. **HTTPS**：生产环境必须使用 HTTPS
5. **CORS 配置**：正确配置跨域策略

---

## 部署建议

### CDN 部署（游戏静态文件）

```
游戏文件 → CDN (加速访问)
    ↓
集成商用户 → 就近的 CDN 节点
```

### 后端部署

```
容器化部署 → Docker
    ↓
负载均衡 → Nginx/Kubernetes
    ↓
多实例 → 水平扩展
```

---

## 常见问题

### Q: 如何处理多币种？

A: 在 setUser 时传入币种信息，后端根据币种处理金额计算。

### Q: 如何支持多语言？

A: 游戏加载时传入语言参数，SDK 根据 `lang` 参数加载对应语言包。

### Q: 离线模式支持吗？

A: Godot Web 版本需要在线加载，不支持完全离线。可考虑使用 Service Worker 缓存。

### Q: 移动端适配？

A: Godot Web 版本支持移动端浏览器，需在导出时配置移动端优化选项。
