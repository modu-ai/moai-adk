// Package normal은 테스트 assertion이 없는 일반 코드 예시를 제공합니다.
// Package normal provides an example of code with no test assertions to lock.
package normal_test

import "testing"

// TestNothing는 assertion이 없어서 lockset 항목이 생성되지 않습니다.
func TestNothing(t *testing.T) {
	t.Log("this test has no string assertions")
}
