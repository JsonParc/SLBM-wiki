package markup

import "testing"

func TestCompat_html_safe_url_iframe_minigame(t *testing.T) {
	allowed := []string{
		"/views/minigame/missile-simulator/index.html",
		"https://www.youtube.com/embed/abc",
	}
	for _, value := range allowed {
		if Compat_html_safe_url(value, true) != value {
			t.Errorf("expected %q to be allowed", value)
		}
	}

	blocked := []string{
		"/views/main_css/js/main.js",
		"/views/minigame/../../w/test",
		"/views/minigame/%2e%2e/x",
		"/views/minigame\\..\\x",
		"//evil.example/views/minigame/x",
		"https://evil.example/views/minigame/x",
		"javascript:alert(1)",
		"views/minigame/missile-simulator/index.html",
	}
	for _, value := range blocked {
		if Compat_html_safe_url(value, true) != "" {
			t.Errorf("expected %q to be blocked", value)
		}
	}
}
