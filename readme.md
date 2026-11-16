
Golang luanti colormappings

![](https://github.com/luanti-golang/colormapping/workflows/test/badge.svg)

# Example

```golang
m := colormapping.NewColorMapping()
err := m.LoadDefaults()
assert.NoError(t, err)

c := m.GetColor("scifi_nodes:blacktile2", 0)
assert.NotNil(t, c)
assert.Equal(t, uint8(20), c.R)
assert.Equal(t, uint8(20), c.G)
assert.Equal(t, uint8(20), c.B)
```

# License

Assets:
* `colors/unifieddyes_palette_extended.png` GPL 2.0 or higher (https://github.com/mt-mods/unifieddyes)

Code: **MIT**