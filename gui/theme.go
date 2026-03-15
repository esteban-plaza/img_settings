package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// appleTheme is a minimal override on top of the default Fyne theme that
// brings the palette and spacing closer to Apple's Human Interface Guidelines.
type appleTheme struct{}

var _ fyne.Theme = appleTheme{}

func (appleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x1c, G: 0x1c, B: 0x1e, A: 0xff} // iOS/macOS dark bg
		}
		return color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf7, A: 0xff} // Apple light bg

	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		}
		return color.NRGBA{R: 0x1d, G: 0x1d, B: 0x1f, A: 0xff}

	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x00, G: 0x71, B: 0xe3, A: 0xff} // Apple blue

	case theme.ColorNameHover:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x14}
		}
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0a}

	case theme.ColorNameButton:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0} // flat/transparent buttons

	case theme.ColorNameInputBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x2c, G: 0x2c, B: 0x2e, A: 0xff}
		}
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	case theme.ColorNameSeparator:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x28}
		}
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x20}
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (appleTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (appleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (appleTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameText:
		return 13 // matches macOS default body text
	}
	return theme.DefaultTheme().Size(name)
}
