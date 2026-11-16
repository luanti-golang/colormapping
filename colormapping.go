package colormapping

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

//go:embed colors/*
var Files embed.FS

type ColorMapping struct {
	// colors holds the explicitly loaded name->color pairs, guarded by mutex
	colors               map[string]*color.RGBA
	extendedpaletteblock map[string]bool
	extendedpalette      *Palette
	mutex                *sync.RWMutex
	// cache holds resolved lookups (including misses as nil) for lock-free reads,
	// it is replaced with an empty map whenever new colors are loaded
	cache atomic.Pointer[sync.Map]
}

type simpleNamedColor struct {
	color                                  *color.RGBA
	darkPrefix, darkSuffix, prefix, suffix string
}

func newSimpleNamedColor(name string, c *color.RGBA) simpleNamedColor {
	return simpleNamedColor{
		color:      c,
		darkPrefix: "dark_" + name,
		darkSuffix: name + "_dark",
		prefix:     name + "_",
		suffix:     "_" + name,
	}
}

// fixed order: the first match wins, so results are deterministic
var simpleNamedColors = []simpleNamedColor{
	newSimpleNamedColor("black", &color.RGBA{R: 32, G: 32, B: 32, A: 255}),
	newSimpleNamedColor("blue", &color.RGBA{R: 59, G: 68, B: 93, A: 255}),
	newSimpleNamedColor("brown", &color.RGBA{R: 48, G: 39, B: 31, A: 255}),
	newSimpleNamedColor("cyan", &color.RGBA{R: 60, G: 96, B: 100, A: 255}),
	newSimpleNamedColor("gray", &color.RGBA{R: 150, G: 148, B: 149, A: 255}),
	newSimpleNamedColor("green", &color.RGBA{R: 93, G: 112, B: 70, A: 255}),
	newSimpleNamedColor("grey", &color.RGBA{R: 150, G: 148, B: 149, A: 255}),
	newSimpleNamedColor("magenta", &color.RGBA{R: 118, G: 93, B: 119, A: 255}),
	newSimpleNamedColor("orange", &color.RGBA{R: 154, G: 116, B: 75, A: 255}),
	newSimpleNamedColor("pink", &color.RGBA{R: 183, G: 150, B: 160, A: 255}),
	newSimpleNamedColor("red", &color.RGBA{R: 107, G: 54, B: 53, A: 255}),
	newSimpleNamedColor("violet", &color.RGBA{R: 70, G: 52, B: 83, A: 255}),
	newSimpleNamedColor("white", &color.RGBA{R: 231, G: 223, B: 225, A: 255}),
	newSimpleNamedColor("yellow", &color.RGBA{R: 155, G: 136, B: 75, A: 255}),
}

func maxUint8(i1, i2 uint8) uint8 {
	if i1 > i2 {
		return i1
	} else {
		return i2
	}
}

func (m *ColorMapping) GetColor(name string, param2 int) *color.RGBA {
	//TODO: list of node->palette
	if m.extendedpaletteblock[name] {
		// param2 coloring
		return m.extendedpalette.GetColor(param2)
	}

	// load the cache once, a concurrent LoadBytes may swap it out (results stored
	// into the old one are discarded with it)
	cache := m.cache.Load()
	if v, ok := cache.Load(name); ok {
		// cached hit or miss
		return v.(*color.RGBA)
	}

	c := m.resolveColor(name)
	cache.Store(name, c)
	return c
}

// resolveColor looks up the color of a node name without consulting the cache,
// returns nil if no color can be determined
func (m *ColorMapping) resolveColor(name string) *color.RGBA {
	m.mutex.RLock()
	c := m.colors[name]
	m.mutex.RUnlock()
	if c != nil {
		// perfect match found
		return c
	}

	// fall back to simple color-name matching
	for _, sc := range simpleNamedColors {
		if strings.Contains(name, sc.darkPrefix) || strings.Contains(name, sc.darkSuffix) {
			// dark color
			return &color.RGBA{
				R: maxUint8(sc.color.R-30, 0),
				G: maxUint8(sc.color.G-30, 0),
				B: maxUint8(sc.color.B-30, 0),
				A: sc.color.A,
			}
		} else if strings.Contains(name, sc.prefix) || strings.Contains(name, sc.suffix) {
			// normal color
			return sc.color
		}
	}

	return nil
}

func (m *ColorMapping) GetColors() map[string]*color.RGBA {
	return m.colors
}

func (m *ColorMapping) LoadBytes(buffer []byte) (int, error) {
	// invalidate cached lookups (including cached misses) after loading
	defer m.cache.Store(&sync.Map{})

	scanner := bufio.NewScanner(bytes.NewReader(buffer))
	count := 0
	line := 0

	for scanner.Scan() {
		line++

		txt := strings.Trim(scanner.Text(), " ")

		if len(txt) == 0 {
			//empty
			continue
		}

		if strings.HasPrefix(txt, "#") {
			//comment
			continue
		}

		parts := strings.Fields(txt)

		if len(parts) < 4 {
			return 0, errors.New("invalid line: #" + strconv.Itoa(line))
		}

		if len(parts) >= 4 {
			r, err := strconv.ParseInt(parts[1], 10, 32)
			if err != nil {
				return 0, fmt.Errorf("red color component parse error '%s': %v", parts[1], err)
			}

			g, err := strconv.ParseInt(parts[2], 10, 32)
			if err != nil {
				return 0, fmt.Errorf("green color component parse error '%s': %v", parts[2], err)
			}

			b, err := strconv.ParseInt(parts[3], 10, 32)
			if err != nil {
				return 0, fmt.Errorf("blue color component parse error '%s': %v", parts[3], err)
			}

			a := int64(255)
			if len(parts) > 4 {
				a, err = strconv.ParseInt(parts[4], 10, 32)
				if err != nil {
					return 0, fmt.Errorf("alpha channer component parse error '%s': %v", parts[4], err)
				}
			}
			c := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
			m.mutex.Lock()
			m.colors[parts[0]] = &c
			m.mutex.Unlock()
			count++
		}
	}

	return count, nil
}

func (m *ColorMapping) LoadVFSColors(filename string) (int, error) {
	buffer, err := Files.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	return m.LoadBytes(buffer)
}

func (m *ColorMapping) LoadDefaults() error {
	list := []string{
		"advtrains.txt",
		"custom.txt",
		"everness.txt",
		"mc2.txt",
		"miles.txt",
		"naturalbiomes.txt",
		"mtg.txt",
		"nodecore.txt",
		"scifi_nodes.txt",
		"vanessa.txt",
	}
	for _, name := range list {
		_, err := m.LoadVFSColors("colors/" + name)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewColorMapping() *ColorMapping {
	data, err := Files.ReadFile("colors/unifieddyes_palette_extended.png")
	if err != nil {
		panic(err)
	}

	extendedpalette, err := NewPalette(data)
	if err != nil {
		panic(err)
	}

	data, err = Files.ReadFile("colors/extended_palette.txt")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	extendedpaletteblock := make(map[string]bool)

	for scanner.Scan() {
		txt := strings.Trim(scanner.Text(), " ")

		if len(txt) == 0 {
			//empty
			continue
		}

		extendedpaletteblock[txt] = true
	}

	m := &ColorMapping{
		colors:               make(map[string]*color.RGBA),
		extendedpaletteblock: extendedpaletteblock,
		extendedpalette:      extendedpalette,
		mutex:                &sync.RWMutex{},
	}
	m.cache.Store(&sync.Map{})
	return m
}
