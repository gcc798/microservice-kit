package runtimeconfig

import "testing"

func TestValidate(t *testing.T) {
	if Validate(CodeWeChat, `{"enabled":true,"appId":"","secret":""}`) == nil {
		t.Fatal("enabled WeChat accepted missing credentials")
	}
	if err := Validate(CodeCaptcha, `{"image":{"enabled":true,"length":4,"width":120,"height":40,"expire":300},"sms":{"enabled":false,"length":6,"expire":300,"template":"x","provider":"x"},"email":{"enabled":false,"length":6,"expire":300,"template":"x"}}`); err != nil {
		t.Fatal(err)
	}
}
