package css3

import (
	"bytes"
)

var BasicColorKeywords = map[string]string{
	"black":   "#000000",
	"silver":  "#c0c0c0",
	"gray":    "#808080",
	"white":   "#ffffff",
	"maroon":  "#800000",
	"red":     "#ff0000",
	"purple":  "#800080",
	"fuchsia": "#ff00ff",
	"green":   "#008000",
	"lime":    "#00ff00",
	"olive":   "#808000",
	"yellow":  "#ffff00",
	"navy":    "#000080",
	"blue":    "#0000ff",
	"teal":    "#008080",
	"aqua":    "#00ffff",
}

var ExtendedColorKeywords = map[string]string{
	"aliceblue":            "#f0f8ff",
	"antiquewhite":         "#faebd7",
	"aqua":                 "#00ffff",
	"aquamarine":           "#7fffd4",
	"azure":                "#f0ffff",
	"beige":                "#f5f5dc",
	"bisque":               "#ffe4c4",
	"black":                "#000000",
	"blanchedalmond":       "#ffebcd",
	"blue":                 "#0000ff",
	"blueviolet":           "#8a2be2",
	"brown":                "#a52a2a",
	"burlywood":            "#deb887",
	"cadetblue":            "#5f9ea0",
	"chartreuse":           "#7fff00",
	"chocolate":            "#d2691e",
	"coral":                "#ff7f50",
	"cornflowerblue":       "#6495ed",
	"cornsilk":             "#fff8dc",
	"crimson":              "#dc143c",
	"cyan":                 "#00ffff",
	"darkblue":             "#00008b",
	"darkcyan":             "#008b8b",
	"darkgoldenrod":        "#b8860b",
	"darkgray":             "#a9a9a9",
	"darkgreen":            "#006400",
	"darkgrey":             "#a9a9a9",
	"darkkhaki":            "#bdb76b",
	"darkmagenta":          "#8b008b",
	"darkolivegreen":       "#556b2f",
	"darkorange":           "#ff8c00",
	"darkorchid":           "#9932cc",
	"darkred":              "#8b0000",
	"darksalmon":           "#e9967a",
	"darkseagreen":         "#8fbc8f",
	"darkslateblue":        "#483d8b",
	"darkslategray":        "#2f4f4f",
	"darkslategrey":        "#2f4f4f",
	"darkturquoise":        "#00ced1",
	"darkviolet":           "#9400d3",
	"deeppink":             "#ff1493",
	"deepskyblue":          "#00bfff",
	"dimgray":              "#696969",
	"dimgrey":              "#696969",
	"dodgerblue":           "#1e90ff",
	"firebrick":            "#b22222",
	"floralwhite":          "#fffaf0",
	"forestgreen":          "#228b22",
	"fuchsia":              "#ff00ff",
	"gainsboro":            "#dcdcdc",
	"ghostwhite":           "#f8f8ff",
	"gold":                 "#ffd700",
	"goldenrod":            "#daa520",
	"gray":                 "#808080",
	"green":                "#008000",
	"greenyellow":          "#adff2f",
	"grey":                 "#808080",
	"honeydew":             "#f0fff0",
	"hotpink":              "#ff69b4",
	"indianred":            "#cd5c5c",
	"indigo":               "#4b0082",
	"ivory":                "#fffff0",
	"khaki":                "#f0e68c",
	"lavender":             "#e6e6fa",
	"lavenderblush":        "#fff0f5",
	"lawngreen":            "#7cfc00",
	"lemonchiffon":         "#fffacd",
	"lightblue":            "#add8e6",
	"lightcoral":           "#f08080",
	"lightcyan":            "#e0ffff",
	"lightgoldenrodyellow": "#fafad2",
	"lightgray":            "#d3d3d3",
	"lightgreen":           "#90ee90",
	"lightgrey":            "#d3d3d3",
	"lightpink":            "#ffb6c1",
	"lightsalmon":          "#ffa07a",
	"lightseagreen":        "#20b2aa",
	"lightskyblue":         "#87cefa",
	"lightslategray":       "#778899",
	"lightslategrey":       "#778899",
	"lightsteelblue":       "#b0c4de",
	"lightyellow":          "#ffffe0",
	"lime":                 "#00ff00",
	"limegreen":            "#32cd32",
	"linen":                "#faf0e6",
	"magenta":              "#ff00ff",
	"maroon":               "#800000",
	"mediumaquamarine":     "#66cdaa",
	"mediumblue":           "#0000cd",
	"mediumorchid":         "#ba55d3",
	"mediumpurple":         "#9370db",
	"mediumseagreen":       "#3cb371",
	"mediumslateblue":      "#7b68ee",
	"mediumspringgreen":    "#00fa9a",
	"mediumturquoise":      "#48d1cc",
	"mediumvioletred":      "#c71585",
	"midnightblue":         "#191970",
	"mintcream":            "#f5fffa",
	"mistyrose":            "#ffe4e1",
	"moccasin":             "#ffe4b5",
	"navajowhite":          "#ffdead",
	"navy":                 "#000080",
	"oldlace":              "#fdf5e6",
	"olive":                "#808000",
	"olivedrab":            "#6b8e23",
	"orange":               "#ffa500",
	"orangered":            "#ff4500",
	"orchid":               "#da70d6",
	"palegoldenrod":        "#eee8aa",
	"palegreen":            "#98fb98",
	"paleturquoise":        "#afeeee",
	"palevioletred":        "#db7093",
	"papayawhip":           "#ffefd5",
	"peachpuff":            "#ffdab9",
	"peru":                 "#cd853f",
	"pink":                 "#ffc0cb",
	"plum":                 "#dda0dd",
	"powderblue":           "#b0e0e6",
	"purple":               "#800080",
	"red":                  "#ff0000",
	"rosybrown":            "#bc8f8f",
	"royalblue":            "#4169e1",
	"saddlebrown":          "#8b4513",
	"salmon":               "#fa8072",
	"sandybrown":           "#f4a460",
	"seagreen":             "#2e8b57",
	"seashell":             "#fff5ee",
	"sienna":               "#a0522d",
	"silver":               "#c0c0c0",
	"skyblue":              "#87ceeb",
	"slateblue":            "#6a5acd",
	"slategray":            "#708090",
	"slategrey":            "#708090",
	"snow":                 "#fffafa",
	"springgreen":          "#00ff7f",
	"steelblue":            "#4682b4",
	"tan":                  "#d2b48c",
	"teal":                 "#008080",
	"thistle":              "#d8bfd8",
	"tomato":               "#ff6347",
	"turquoise":            "#40e0d0",
	"violet":               "#ee82ee",
	"wheat":                "#f5deb3",
	"white":                "#ffffff",
	"whitesmoke":           "#f5f5f5",
	"yellow":               "#ffff00",
	"yellowgreen":          "#9acd32",
}

type Color struct {
	R            float64
	G            float64
	B            float64
	A            float64
	CurrentColor bool
}

func RGB(r, g, b float64) *Color {
	return &Color{r, g, b, 1, false}
}

func RGBA(r, g, b, a float64) *Color {
	return &Color{r, g, b, clamp(a), false}
}

func HSL(h, s, v float64) *Color {
	return HSLA(h, s, v, 1)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func hue(m1, m2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	if h*6 < 1 {
		return m1 + (m2-m1)*h*6
	}
	if h*2 < 1 {
		return m2
	}
	if h*3 < 2 {
		return m1 + (m2-m1)*(2.0/3-h)*6
	}
	return m1
}

func normDeg(deg float64) float64 {
	return float64(((int(deg)%360)+360)%360) / 360
}

func HSLA(h, s, l, a float64) *Color {
	h = clamp(h)
	s = clamp(s)
	l = clamp(l)
	a = clamp(a)
	var m2 float64
	if l <= 0.5 {
		m2 = l * (s + 1)
	} else {
		m2 = l + s - l*s
	}
	m1 := l*2 - m2
	return &Color{
		R: hue(m1, m2, h+1.0/3),
		G: hue(m1, m2, h),
		B: hue(m1, m2, h-1.0/3),
		A: a,
	}
}

func ColorFromString(s string) *Color {
	parser := NewParser(bytes.NewReader([]byte(s)))
	nodes := parser.ParseListOfComponentValues()
	return ColorFromNodes(nodes)
}

func ColorFromNodes(nodes []Node) *Color {
	for len(nodes) > 0 {
		if n, ok := nodes[0].(*TokenNode); ok && n.TokenType == WhitespaceToken {
			nodes = nodes[1:]
		} else {
			break
		}
	}
	if len(nodes) == 0 {
		return nil
	}
	switch n := nodes[0].(type) {
	case *FunctionNode:
		return n.Color()
	case *HashNode:
		return ColorFromHexCode(n.Hash)
	case *TokenNode:
		if n.TokenType == IdentToken {
			return ColorFromName(string(n.Value.(Identifier)))
		}
		return nil
	default:
		return nil
	}
}

func ColorFromName(name string) *Color {
	name = toLower(name)
	if name == "transparent" {
		return new(Color)
	}
	if name == "currentcolor" {
		return &Color{CurrentColor: true}
	}
	if code, ok := BasicColorKeywords[name]; ok {
		return ColorFromHexCode(code[1:])
	}
	if code, ok := ExtendedColorKeywords[name]; ok {
		return ColorFromHexCode(code[1:])
	}
	return nil
}

func hexCodeToColorComponent(code string) float64 {
	var x, y int
	x = parseHexDigit(rune(code[0]))
	if len(code) > 1 {
		y = parseHexDigit(rune(code[1]))
	} else {
		y = x
	}
	if x < 0 || y < 0 {
		return -1
	}
	return float64(x*16+y) / 255
}

func ColorFromHexCode(code string) (color *Color) {
	if len(code) == 3 {
		color = &Color{
			R: hexCodeToColorComponent(code[0:1]),
			G: hexCodeToColorComponent(code[1:2]),
			B: hexCodeToColorComponent(code[2:3]),
			A: 1.0,
		}
	} else if len(code) == 6 {
		color = &Color{
			R: hexCodeToColorComponent(code[0:2]),
			G: hexCodeToColorComponent(code[2:4]),
			B: hexCodeToColorComponent(code[4:6]),
			A: 1.0,
		}
	} else {
		return nil
	}
	if color.R < 0 || color.G < 0 || color.B < 0 {
		return nil
	}
	return
}

func (c *Color) TestRepr() interface{} {
	if c == nil {
		return nil
	}
	if c.CurrentColor {
		return "currentColor"
	}
	return []float64{c.R, c.G, c.B, c.A}
}
