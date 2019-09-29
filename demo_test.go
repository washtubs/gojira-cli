// +build demo

package cli

import "testing"

func TestFzfBasic(t *testing.T) {
	demoApp := NewApp()
	i, err := demoApp.actionBaseService.BuildAction()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(i.Format())
}
