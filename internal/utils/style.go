package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

var (
	BlackStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "0",
		Dark:  "8",
	})
	RedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "1",
		Dark:  "9",
	})
	GreenStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "2",
		Dark:  "10",
	})
	YellowStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "3",
		Dark:  "11",
	})
	BlueStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "4",
		Dark:  "12",
	})
	MagentaStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "5",
		Dark:  "13",
	})
	CyanStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "6",
		Dark:  "14",
	})
	WhiteStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "7",
		Dark:  "15",
	})

	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	TableRowStyle    = lipgloss.NewStyle().Padding(0, 1)
)

var (
	ErrorPrefix   = RedStyle.Bold(true).Render("Error: ")
	WarningPrefix = YellowStyle.Bold(true).Render("Warning: ")
)

func PrintError(err error) {
	fmt.Println(ErrorPrefix + RedStyle.Render(err.Error()))
}

func PrintErrorWithMessage(message string, err error) {
	if err != nil {
		message = message + ": " + err.Error()
	}
	fmt.Println(ErrorPrefix + RedStyle.Render(message))
}

func PrintWarning(message string) {
	fmt.Println(WarningPrefix + YellowStyle.Render(message))
}

func PrintSuccess(message string) {
	fmt.Println(GreenStyle.Bold(true).Render(message))
}

var (
	TestFunctionResultStyle = TestFunctionResult{
		EmptyText:  "<empty>",
		EmptyStyle: BlackStyle,
		ErrorStyle: RedStyle,
		ValueStyle: GreenStyle,
	}

	TestFunctionResultListStyle = TestFunctionResult{
		EmptyText:  "<empty>",
		EmptyStyle: BlackStyle,
		ErrorStyle: RedStyle,
		ValueStyle: GreenStyle,
	}

	MaybeEnabledStyle = MaybeEnabled{
		TrueText:             SymCheckMark,
		TrueStyle:            GreenStyle.Bold(true),
		FalseText:            SymCrossMark,
		FalseStyle:           RedStyle.Bold(true),
		InheritedFalsePrefix: "",
		InheritedFalseSuffix: "*",
		InheritedFalseStyle:  RedStyle.Bold(true),
		NullText:             "-",
		NullStyle:            BlackStyle,
	}

	MaybeEnabledTableStyle = MaybeEnabled{
		TrueText:             SymCheckMark,
		TrueStyle:            GreenStyle.Bold(true),
		FalseText:            SymCrossMark,
		FalseStyle:           RedStyle.Bold(true),
		InheritedFalsePrefix: " ",
		InheritedFalseSuffix: "*",
		InheritedFalseStyle:  RedStyle.Bold(true),
		NullText:             "-",
		NullStyle:            BlackStyle,
	}

	MaybeEmptyStyle = MaybeEmpty{
		EmptyText:  "<empty>",
		EmptyStyle: BlackStyle,
		NullText:   "-",
		NullStyle:  BlackStyle,
	}

	MaybeWildcardNameStyle = MaybeEmpty{
		NullText:  "%",
		NullStyle: CyanStyle.Bold(true),
	}

	MaybeZeroStyle = MaybeZero{
		ZeroText:  "0",
		ZeroStyle: YellowStyle.Bold(true),
	}

	MaybeTimeStyle = MaybeTime{
		NullText:  "-",
		NullStyle: BlackStyle,
		Format:    "2006-01-02 15:04:05",
	}

	MaybeIDSuffixStyle = MaybeIDSuffix{
		SuffixStyle: BlueStyle,
		NullText:    "-",
		NullStyle:   BlackStyle,
	}

	MaybePasswordStyle = MaybePassword{
		NullText:    "-",
		NullStyle:   BlackStyle,
		HiddenText:  "***",
		HiddenStyle: BlueStyle.Bold(true),
	}

	MaybeQuotaStyle = MaybeQuota{
		UnlimitedText:  SymInfinity,
		UnlimitedStyle: MagentaStyle.Bold(true),
	}

	SQLLikeStyle = SQLLike{
		WildcardStyle: CyanStyle.Bold(true),
	}
)

type TestFunctionResult struct {
	EmptyText  string
	EmptyStyle lipgloss.Style
	ErrorStyle lipgloss.Style
	ValueStyle lipgloss.Style
}

func (fr TestFunctionResult) Render(result any, err error) string {
	if errors.Is(err, sql.ErrNoRows) {
		return fr.EmptyStyle.Render(fr.EmptyText)
	} else if err != nil {
		return fr.ErrorStyle.Render(err.Error())
	}

	switch v := result.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			return fr.ValueStyle.Render(v)
		} else {
			return fr.EmptyStyle.Render(fr.EmptyText)
		}
	case []string:
		if len(v) == 0 {
			return fr.EmptyStyle.Render(fr.EmptyText)
		} else {
			l := list.New()
			l.Items(v)
			return l.String()
		}
	default:
		panic("unsupported type for TestFunctionResult.Render")
	}
}

type MaybeEnabled struct {
	TrueText             string
	TrueStyle            lipgloss.Style
	FalseText            string
	FalseStyle           lipgloss.Style
	InheritedFalsePrefix string
	InheritedFalseSuffix string
	InheritedFalseStyle  lipgloss.Style
	NullText             string
	NullStyle            lipgloss.Style
}

func (e MaybeEnabled) Render(value any, inherited ...any) string {
	enabledVal, ok := ToBoolPtr(value)
	if !ok {
		panic("unsupported type for MaybeEnabled.Render")
	}

	var inheritedFalse bool
	for _, iv := range inherited {
		switch ivt := iv.(type) {
		case bool:
			if !ivt {
				inheritedFalse = true
				break
			}
		case *bool:
			if ivt != nil && !*ivt {
				inheritedFalse = true
				break
			}
		case sql.NullBool:
			if ivt.Valid && !ivt.Bool {
				inheritedFalse = true
				break
			}
		default:
			panic("unsupported type for MaybeEnabled.Render inherited")
		}
	}

	var prefix, suffix string
	if inheritedFalse {
		prefix = e.InheritedFalseStyle.Render(e.InheritedFalsePrefix)
		suffix = e.InheritedFalseStyle.Render(e.InheritedFalseSuffix)
	}

	out := prefix
	if enabledVal == nil {
		out += e.NullStyle.Render(e.NullText)
	} else if *enabledVal {
		out += e.TrueStyle.Render(e.TrueText)
	} else {
		out += e.FalseStyle.Render(e.FalseText)
	}
	out += suffix
	return out
}

type MaybeEmpty struct {
	EmptyText  string
	EmptyStyle lipgloss.Style
	NullText   string
	NullStyle  lipgloss.Style
}

func (me MaybeEmpty) Render(value any) string {
	var ok bool
	var strVal *string
	var boolVal *bool
	var uintVal *uint64

	if value == nil {
		goto render
	}

	strVal, ok = ToStringPtr(value)
	if ok {
		goto render
	}

	boolVal, ok = ToBoolPtr(value)
	if ok {
		if boolVal != nil {
			str := fmt.Sprintf("%t", *boolVal)
			strVal = &str
		}
		goto render
	}

	uintVal, ok = ToUint64Ptr(value)
	if ok {
		if uintVal != nil {
			str := fmt.Sprintf("%d", *uintVal)
			strVal = &str
		}
		goto render
	}

	panic("unsupported type for MaybeEmpty.Render")

render:
	if strVal == nil {
		return me.NullStyle.Render(me.NullText)
	} else if strings.TrimSpace(*strVal) == "" {
		return me.EmptyStyle.Render(me.EmptyText)
	} else {
		return *strVal
	}
}

type MaybeZero struct {
	ZeroText  string
	ZeroStyle lipgloss.Style
}

func (mz MaybeZero) Render(value any) string {
	if value == 0 {
		return mz.ZeroStyle.Render(mz.ZeroText)
	}
	return fmt.Sprintf("%d", value)
}

type MaybeTime struct {
	NullText  string
	NullStyle lipgloss.Style
	Format    string
}

func (mt MaybeTime) Render(value any) string {
	timeVal, ok := ToTimePtr(value)
	if !ok {
		panic("unsupported type for MaybeTime.Render")
	}

	if timeVal == nil {
		return mt.NullStyle.Render(mt.NullText)
	} else {
		return timeVal.Format(mt.Format)
	}
}

type MaybeIDSuffix struct {
	SuffixStyle lipgloss.Style
	NullText    string
	NullStyle   lipgloss.Style
}

func (ids MaybeIDSuffix) Render(str any, suffix any) string {
	strVal, ok := ToStringPtr(str)
	if !ok {
		panic("unsupported type for MaybeIDSuffix.Render")
	}

	var suffixVal *string
	suffixVal, ok = ToStringPtr(suffix)
	if !ok {
		panic("unsupported type for MaybeIDSuffix.Render suffix")
	}

	if strVal == nil {
		return ids.NullStyle.Render(ids.NullText)
	}

	out := *strVal
	if suffixVal != nil {
		out = out + " " + ids.SuffixStyle.Render("("+*suffixVal+")")
	}
	return out
}

type MaybePassword struct {
	NullText    string
	NullStyle   lipgloss.Style
	HiddenText  string
	HiddenStyle lipgloss.Style
}

func (mp MaybePassword) Render(isSet any) string {
	isSetVal, ok := ToBoolPtr(isSet)
	if !ok {
		panic("unsupported type for MaybePassword.Render")
	}

	if isSetVal == nil || !*isSetVal {
		return mp.NullStyle.Render(mp.NullText)
	} else {
		return mp.HiddenStyle.Render(mp.HiddenText)
	}
}

type MaybeQuota struct {
	UnlimitedText  string
	UnlimitedStyle lipgloss.Style
}

func (mq MaybeQuota) Render(quotaBytes any, scaling uint64) string {
	quotaVal, ok := ToUint64Ptr(quotaBytes)
	if !ok {
		panic("unsupported type for MaybeQuota.Render")
	}

	if quotaVal == nil || *quotaVal == 0 {
		return mq.UnlimitedStyle.Render(mq.UnlimitedText)
	}

	return FormatBytes(*quotaVal * scaling)
}

type SQLLike struct {
	WildcardStyle lipgloss.Style
}

func (sl SQLLike) Render(str string) string {
	out := ""
	for _, r := range str {
		if r == '%' || r == '_' {
			out += sl.WildcardStyle.Render(string(r))
		} else {
			out += string(r)
		}
	}
	return out
}
