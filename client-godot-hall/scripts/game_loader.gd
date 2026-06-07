extends Node

## 子游戏加载器 - 负责加载、卸载和管理子游戏实例
## 使用独立的SubViewport来隔离子游戏渲染

signal game_loaded(game_node: Node, game_id: String)
signal game_unloaded(game_id: String)
signal load_failed(error: String)

var _game_viewport: SubViewport = null
var _current_game: Node = null
var _current_game_id: String = ""

## 设置目标视口
func set_viewport(viewport: SubViewport) -> void:
    _game_viewport = viewport

## 加载游戏
func load_game(game_id: String) -> void:
    if not _game_viewport:
        _emit_error("游戏视口未设置")
        return

    if _current_game:
        _emit_error("已有游戏运行中，请先卸载")
        return

    var game_info = SubGameRegistry.get_game_info(game_id)
    if not game_info:
        _emit_error("游戏不存在: %s" % game_id)
        return

    print("[GameLoader] 开始加载游戏: %s (场景: %s)" % [game_id, game_info.scene_path])

    if not ResourceLoader.exists(game_info.scene_path):
        _emit_error("场景文件不存在: %s" % game_info.scene_path)
        return

    await get_tree().process_frame
    var game_scene = load(game_info.scene_path)

    if not game_scene:
        _emit_error("场景加载失败: %s" % game_info.scene_path)
        return

    _current_game = game_scene.instantiate()

    if not _current_game:
        _emit_error("场景实例化失败")
        return

    _game_viewport.add_child(_current_game)
    _current_game_id = game_id

    if _current_game.has_method("on_game_start"):
        _current_game.on_game_start()

    print("[GameLoader] 游戏加载完成: %s" % game_id)
    game_loaded.emit(_current_game, game_id)

## 卸载当前游戏
func unload_game() -> void:
    if not _current_game:
        game_unloaded.emit("")
        return

    print("[GameLoader] 卸载游戏: %s" % _current_game_id)

    if _current_game.has_method("on_game_exit"):
        _current_game.on_game_exit()

    if _current_game.get_parent():
        _current_game.get_parent().remove_child(_current_game)

    _current_game.queue_free()
    var temp_id = _current_game_id
    _current_game = null
    _current_game_id = ""

    await get_tree().process_frame
    game_unloaded.emit(temp_id)

## 获取当前运行的游戏
func get_current_game() -> Node:
    return _current_game

## 获取当前游戏ID
func get_current_game_id() -> String:
    return _current_game_id

## 是否有游戏运行中
func is_game_running() -> bool:
    return _current_game != null

func _emit_error(error: String) -> void:
    printerr("[GameLoader] 错误: %s" % error)
    load_failed.emit(error)
