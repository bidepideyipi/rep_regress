# 大厅子游戏框架

演示使用 Godot 4.2 实现的大厅加载子游戏的逻辑框架。

## 项目结构

```
client-godot-holl/
├── project.godot              # Godot 项目配置
├── scenes/
│   ├── main_lobby.tscn        # 大厅主场景
│   └── sub_games/            # 子游戏目录
│       ├── demo_game_1.tscn  # 示例游戏1
│       └── demo_game_2.tscn  # 示例游戏2
└── scripts/
    ├── main_lobby.gd          # 大厅逻辑
    ├── game_loader.gd         # 子游戏加载器（自动加载）
    ├── game_manager.gd        # 全局游戏状态管理（自动加载）
    ├── sub_game_registry.gd   # 子游戏注册表（自动加载）
    └── sub_games/
        ├── demo_game_1.gd
        └── demo_game_2.gd
```

## 核心组件

### 1. 自动加载单例 (AutoLoad)

- **GameLoader**: 负责子游戏的加载/卸载，使用 SubViewport 隔离子游戏渲染
- **GameManager**: 全局游戏状态管理，处理游戏间数据传递
- **SubGameRegistry**: 子游戏注册表，管理所有可加载的子游戏信息

### 2. 大厅场景 (MainLobby)

- 显示可用游戏列表
- 处理游戏选择和启动
- 使用 SubViewport 展示正在运行的子游戏
- 支持 ESC 键快速卸载游戏返回大厅

### 3. 子游戏接口规范

每个子游戏需要实现以下方法（可选但推荐）：

```gdscript
## 游戏启动时调用
func on_game_start() -> void:
    pass

## 游戏退出时调用
func on_game_exit() -> void:
    pass
```

## 使用方法

### 添加新的子游戏

1. 在 `scenes/sub_games/` 创建游戏场景
2. 在 `scripts/sub_games/` 创建游戏脚本，实现 `on_game_start()` 和 `on_game_exit()`
3. 在 `SubGameRegistry._register_default_games()` 中注册游戏：

```gdscript
register_game("my_game", {
    name = "我的游戏",
    scene_path = "res://scenes/sub_games/my_game.tscn",
    description = "游戏描述",
    author = "作者名"
})
```

### 与全局数据交互

```gdscript
# 保存数据
GameManager.set_player_data("key", value)

# 读取数据
var value = GameManager.get_player_data("key", default_value)
```

## 运行

使用 Godot 4.2+ 打开项目，直接运行即可。

## 特性

- SubViewport 隔离子游戏渲染，避免状态污染
- 自动加载单例全局访问
- 子游戏生命周期管理
- 游戏间数据共享机制
- 支持 ESC 快捷键返回大厅
