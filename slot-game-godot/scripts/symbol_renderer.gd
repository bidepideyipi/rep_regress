extends Control

class_name SymbolRenderer

static var _instance: SymbolRenderer = null
var symbol_colors: Dictionary = {}

const SYMBOL_SIZE = 100

static func get_instance() -> SymbolRenderer:
	if _instance == null:
		_instance = SymbolRenderer.new()
	return _instance

func create_symbol(symbol_type: String, size: int = SYMBOL_SIZE) -> Control:
	var container = Control.new()
	container.custom_minimum_size = Vector2(size, size)

	var texture_path = "res://assets/symbols/" + symbol_type + ".png"
	var texture = load(texture_path)

	if texture != null:
		var sprite = TextureRect.new()
		sprite.texture = texture
		sprite.custom_minimum_size = Vector2(size, size)
		sprite.expand_mode = TextureRect.EXPAND_FIT_WIDTH_PROPORTIONAL
		sprite.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
		container.add_child(sprite)
		print("Loaded image: ", texture_path)
		return container

	#print("Image not found: ", texture_path, ", using emoji fallback")
	return create_emoji_symbol(symbol_type, size)

func create_emoji_symbol(symbol_type: String, size: int) -> Control:
	var container = Control.new()
	container.custom_minimum_size = Vector2(size, size)

	var bg = ColorRect.new()
	bg.color = Color(0.1, 0.1, 0.15)
	bg.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	container.add_child(bg)

	var panel = Panel.new()
	panel.custom_minimum_size = Vector2(size, size)
	panel.mouse_filter = Control.MOUSE_FILTER_IGNORE
	container.add_child(panel)

	var symbol_label = Label.new()
	symbol_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	symbol_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	symbol_label.add_theme_font_size_override("font_size", 48)

	var emoji_map = {
		"cherry": "🍒",
		"lemon": "🍋",
		"orange": "🍊",
		"grape": "🍇",
		"plum": "🍑",
		"watermelon": "🍉",
		"bell": "🔔",
		"seven": "7",
		"wild": "Wild",
		"scatter": "Scat"
	}

	symbol_label.text = emoji_map.get(symbol_type, "?")

	if symbol_type == "wild":
		symbol_label.add_theme_font_size_override("font_size", 36)
	elif symbol_type == "scatter":
		symbol_label.add_theme_font_size_override("font_size", 36)

	symbol_label.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	container.add_child(symbol_label)

	return container
