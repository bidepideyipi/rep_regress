# Slot Game - Godot 4

这是一个基于 Godot 4 Engine (GDScript) 的 3x3 老虎机游戏前端，对接后端 slot-game 服务的 spin 接口。

## 项目结构

```
slot-game-godot/
├── project.godot              # Godot 项目配置文件
├── icon.svg                   # 项目图标
├── README.md                  # 项目说明文档
├── scenes/                    # 场景文件目录
│   └── main_game.tscn         # 主游戏场景
├── scripts/                   # 脚本目录
│   ├── main_game.gd           # 主游戏逻辑脚本
│   ├── http_manager.gd        # HTTP请求管理器
│   ├── symbol_renderer.gd    # 符号渲染器
│   └── game_models.gd        # 游戏数据模型
└── assets/                    # 资源目录
    └── symbols/               # 符号资源目录
```

## 功能特性

### 核心功能
- **3x3 老虎机游戏** - 经典的 3 卷轴 3 行老虎机布局
- **Spin 接口对接** - 完整对接后端 slot-game 服务的 spin API
- **动态符号渲染** - 使用 Control 组件和 emoji 显示游戏符号
- **动画效果** - 包含旋转动画、中奖特效等
- **下注系统** - 支持调整下注金额

### 游戏符号
游戏包含以下符号，每种符号都有独特的图形表示：
- 🍒 樱桃 - 红色
- 🍋 柠檬 - 黄色
- 🍊 橘子 - 橙色
- 🍇 李子/葡萄 - 紫色
- 🍉 西瓜 - 绿色
- 🔔 铃铛 - 金色
- 7️⃣ 数字7 - 红色
- WILD - 金色带 W
- SCATTER - 青色带 S

## 依赖要求

### 系统要求
- **Godot Engine 4.3+**
- 现代操作系统 (Windows, macOS, Linux)

### 网络要求
- 后端服务运行在 `http://127.0.0.1:8081`
- 需要支持 HTTP 请求的网络连接

## 快速开始

### 1. 安装 Godot Engine 4

从 [Godot 官网](https://godotengine.org/download) 下载并安装 Godot Engine 4.3 或更高版本。

### 2. 打开项目

1. 启动 Godot Engine
2. 点击 "Import" 按钮
3. 选择 `slot-game-godot` 目录
4. 点击 "Import & Edit"

### 3. 运行游戏

点击 Godot 编辑器右上角的 "Play" 按钮即可运行游戏。

或者按 F5 键直接运行。

## API 接口

游戏会调用以下后端接口：

### Spin 接口
- **URL**: `POST http://127.0.0.1:8081/api/game/spin`
- **请求参数**:
```gdscript
class_name SpinRequest
var user_id: String      # 用户ID
var bet_amount: float    # 下注金额
var bet_lines: int       # 下注线数
var session_id: String   # 会话ID
var is_free_spin: bool   # 是否免费旋转
```

- **响应参数**:
```gdscript
class_name SpinResponse
var success: bool
var message: String
var data: Dictionary     # 包含 reel_result, win_amount 等数据
```

## 使用说明

### 游戏操作

1. **调整下注金额**
   - 点击 `+ Bet` 按钮增加下注金额（每次 +0.5）
   - 点击 `- Bet` 按钮减少下注金额（每次 -0.5）
   - 下注范围：0.5 - 100.0

2. **开始游戏**
   - 点击 `SPIN` 按钮开始旋转
   - 每次旋转消耗当前下注金额
   - 初始余额：1000.0

3. **查看结果**
   - 游戏显示旋转后的符号组合
   - 显示本次赢得的金额
   - 显示当前余额

## 代码示例

### 修改后端服务地址

在 `scripts/http_manager.gd` 中修改默认 URL：

```gdscript
var base_url: String = "http://your-server:8081"
```

或运行时动态设置：

```gdscript
HttpManager.get_instance().set_base_url("http://your-server:8081")
```

### 调整游戏参数

在 `scripts/main_game.gd` 中可以调整：

```gdscript
const INITIAL_BALANCE = 1000.0      # 初始余额
const SYMBOL_SIZE = 100            # 符号大小
const REEL_SPACING = 120           # 卷轴间距
const ROW_SPACING = 120            # 行间距

var current_bet: float = 1.0        # 初始下注金额
var bet_lines: int = 5             # 下注线数
var user_id: String = "user_godot_001"  # 用户ID
```

## 自定义符号

如果需要添加新的游戏符号，可以修改 `scripts/symbol_renderer.gd`：

1. 在 `_init_symbols()` 方法中添加符号颜色：
```gdscript
symbol_colors["newsymbol"] = Color(1.0, 0.5, 0.5)
```

2. 在 `create_symbol()` 方法中添加新的符号类型处理：
```gdscript
"newsymbol":
    symbol_label.text = "🎁"
    panel.modulate = Color(1.0, 0.5, 0.5)
```

## 扩展功能

### 添加音效支持

1. 将音频文件放入 `assets/sounds/` 目录
2. 在游戏逻辑中添加音效播放：

```gdscript
var audio_player = AudioStreamPlayer.new()
add_child(audio_player)
audio_player.stream = load("res://assets/sounds/spin.mp3")
audio_player.play()
```

### 添加图片素材支持

修改 `scripts/symbol_renderer.gd`：

```gdscript
func create_symbol(symbol_type: String, size: int = SYMBOL_SIZE) -> Control:
    var texture_path = "res://assets/symbols/" + symbol_type + ".png"
    if ResourceLoader.exists(texture_path):
        var sprite = TextureRect.new()
        sprite.texture = load(texture_path)
        sprite.custom_minimum_size = Vector2(size, size)
        sprite.expand_mode = TextureRect.EXPAND_FIT_WIDTH_PROPORTIONAL
        return sprite
    # 使用 emoji 方式
    ...
```

### 添加粒子效果

```gdscript
func _show_win_particles():
    var particles = CPUParticles2D.new()
    particles.emitting = true
    particles.amount = 50
    particles.lifetime = 2.0
    add_child(particles)
    await get_tree().create_timer(2.0).timeout
    particles.queue_free()
```

## 常见问题

### 问题 1: 网络请求失败
**解决方案:**
- 确保后端服务正在运行
- 检查后端服务地址是否正确
- 检查网络连接和防火墙设置

### 问题 2: 符号显示异常
**解决方案:**
- 确保 SymbolRenderer 正确初始化
- 检查符号名称是否与后端返回的一致
- 查看 Godot 输出面板的错误信息

### 问题 3: 场景加载失败
**解决方案:**
- 检查 project.godot 中的 main_scene 路径是否正确
- 确保所有脚本文件存在
- 重新导入项目

## 性能优化建议

1. **资源预加载**: 在游戏启动时预加载常用资源
2. **对象池**: 对频繁创建销毁的对象使用对象池
3. **帧率控制**: 根据设备性能调整目标帧率
4. **减少重绘**: 只在必要时调用 `queue_redraw()`

## 与其他方案对比

| 特性 | Godot 4 | Cocos Creator | Cocos2d-x |
|------|---------|---------------|-----------|
| 开发语言 | GDScript | TypeScript | C++ |
| 学习曲线 | 低 | 较低 | 较高 |
| 开发效率 | 高 | 高 | 中 |
| 性能 | 优秀 | 良好 | 优秀 |
| 跨平台 | 优秀 | 优秀 | 优秀 |
| 调试便利性 | 优秀 | 优秀 | 一般 |
| 热更新 | 支持 | 支持 | 困难 |
| 引擎大小 | 小 (~100MB) | 中 (~500MB) | 大 |
| 许可证 | MIT (免费) | 商业/免费 | MIT |

## 推荐选择

**选择 Godot 4 如果:**
- 希望快速开发和迭代
- 团队熟悉 GDScript 或愿意学习
- 需要完全免费和开源的方案
- 引擎体积要求小
- 重视开发效率和社区支持

**选择 Cocos Creator 如果:**
- 团队熟悉 JavaScript/TypeScript
- 需要成熟的商业支持
- 主要针对 Web 平台
- 需要丰富的第三方库

**选择 Cocos2d-x 如果:**
- 追求极致性能
- 有 C++ 开发经验
- 针对移动端原生应用
- 需要底层控制

## 构建发布

### Web 平台

1. 在 Godot 编辑器中，点击 `Project` > `Export`
2. 添加 `Web` 平台
3. 配置导出设置
4. 点击 `Export Project`

### 桌面平台

1. 在 Godot 编辑器中，点击 `Project` > `Export`
2. 添加 `Windows Desktop`、`macOS` 或 `Linux/X11` 平台
3. 配置导出设置
4. 点击 `Export Project`

### 移动平台

1. 在 Godot 编辑器中，点击 `Project` > `Export`
2. 添加 `Android` 或 `iOS` 平台
3. 配置导出设置
4. 点击 `Export Project`

## 许可证

本项目仅用于学习和演示目的。

## 贡献

欢迎提交问题和改进建议！

## 联系方式

如有问题，请联系项目维护者。
