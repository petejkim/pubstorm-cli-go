package tui

import "fmt"

const (
	bold      = 1
	underline = 4
	blink     = 5
	invert    = 7

	black   = 0
	red     = 1
	green   = 2
	yellow  = 3
	blue    = 4
	magenta = 5
	cyan    = 6
	white   = 7

	DisableColors = false
)

func modeText(color int, s string) string {
	if DisableColors {
		return s
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, s)
}

func Blk(s string) string { return modeText(30+black, s) }
func Red(s string) string { return modeText(30+red, s) }
func Grn(s string) string { return modeText(30+green, s) }
func Ylo(s string) string { return modeText(30+yellow, s) }
func Blu(s string) string { return modeText(30+blue, s) }
func Mag(s string) string { return modeText(30+magenta, s) }
func Cyn(s string) string { return modeText(30+cyan, s) }
func Wht(s string) string { return modeText(30+white, s) }

func HiBlk(s string) string { return modeText(90+black, s) }
func HiRed(s string) string { return modeText(90+red, s) }
func HiGrn(s string) string { return modeText(90+green, s) }
func HiYlo(s string) string { return modeText(90+yellow, s) }
func HiBlu(s string) string { return modeText(90+blue, s) }
func HiMag(s string) string { return modeText(90+magenta, s) }
func HiCyn(s string) string { return modeText(90+cyan, s) }
func HiWht(s string) string { return modeText(90+white, s) }

func BgBlk(s string) string { return modeText(40+black, s) }
func BgRed(s string) string { return modeText(40+red, s) }
func BgGrn(s string) string { return modeText(40+green, s) }
func BgYlo(s string) string { return modeText(40+yellow, s) }
func BgBlu(s string) string { return modeText(40+blue, s) }
func BgMag(s string) string { return modeText(40+magenta, s) }
func BgCyn(s string) string { return modeText(40+cyan, s) }
func BgWht(s string) string { return modeText(40+white, s) }

func BgHiBlk(s string) string { return modeText(100+black, s) }
func BgHiRed(s string) string { return modeText(100+red, s) }
func BgHiGrn(s string) string { return modeText(100+green, s) }
func BgHiYlo(s string) string { return modeText(100+yellow, s) }
func BgHiBlu(s string) string { return modeText(100+blue, s) }
func BgHiMag(s string) string { return modeText(100+magenta, s) }
func BgHiCyn(s string) string { return modeText(100+cyan, s) }
func BgHiWht(s string) string { return modeText(100+white, s) }

func Bold(s string) string { return modeText(bold, s) }
func Undl(s string) string { return modeText(underline, s) }
func Blnk(s string) string { return modeText(blink, s) }
func Invt(s string) string { return modeText(invert, s) }
