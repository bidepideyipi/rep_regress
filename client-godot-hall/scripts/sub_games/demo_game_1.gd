extends Control

## 示例游戏1 - 演示子游戏的基本结构
## 实现点击计数器和运行时间显示

var click_count: int = 0
var run_time: float = 0.0

@onready var counter_label = $CenterContainer/VBox/Counter
@onready var timer_label = $CenterContainer/VBox/Timer
@onready var click_button = $CenterContainer/VBox/ClickButton

## 游戏启动回调（由GameLoader调用）
func on_game_start() -> void:
    print("[DemoGame1] 游戏开始")
    click_count = 0
    run_time = 0.0
    _update_display()

    click_button.pressed.connect(_on_click)

## 游戏退出回调（由GameLoader调用）
func on_game_exit() -> void:
    print("[DemoGame1] 游戏退出，最终点击数: %d" % click_count)

## 点击按钮处理
func _on_click() -> void:
    click_count += 1
    _update_display()
    print("[DemoGame1] 点击计数: %d" % click_count)

## 更新显示
func _update_display() -> void:
    counter_label.text = "点击计数: %d" % click_count

func _process(delta: float) -> void:
    run_time += delta
    timer_label.text = "运行时间: %.1fs" % run_time
