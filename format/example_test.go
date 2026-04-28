package format

import "fmt"

func ExampleDuration() {
	fmt.Println(Duration(30))
	fmt.Println(Duration(90))
	fmt.Println(Duration(3600))
	// Output:
	// 30s
	// 1m 30s
	// 1h
}

func ExampleDuration_hoursAndMinutes() {
	fmt.Println(Duration(3*3600 + 45*60))
	// Output:
	// 3h 45m
}

func ExampleCost() {
	small := 0.005
	normal := 3.42
	zero := 0.0
	fmt.Println(Cost(nil))
	fmt.Println(Cost(&zero))
	fmt.Println(Cost(&small))
	fmt.Println(Cost(&normal))
	// Output:
	// --
	// --
	// $0.0050
	// $3.42
}

func ExampleTokens() {
	small := 999
	large := 1500
	fmt.Println(Tokens(nil))
	fmt.Println(Tokens(&small))
	fmt.Println(Tokens(&large))
	// Output:
	// --
	// 999
	// 1.5k
}

func ExampleProviderName() {
	name := "openai"
	fmt.Println(ProviderName(nil))
	fmt.Println(ProviderName(&name))
	// Output:
	// --
	// openai
}

func ExampleRelativeTime() {
	// RelativeTime depends on the wall clock, so this example is
	// compile-only; pkg.go.dev will render it without a verified output.
	_ = RelativeTime("2026-01-01T00:00:00Z")
}

func ExampleTimestamp() {
	// Timestamp formats in the caller's local time zone, so this example
	// is compile-only; pkg.go.dev will render it without a verified output.
	_ = Timestamp("2026-01-01T12:34:56Z")
}

func ExampleCapacityRatio() {
	max := 8
	fmt.Println(CapacityRatio(5, &max))
	fmt.Println(CapacityRatio(5, nil))
	// Output:
	// 5 / 8
	// 5 / ∞
}

func ExampleAttestationFingerprint() {
	fmt.Println(AttestationFingerprint("ed25519:abc1234abcdd4f2"))
	fmt.Println(AttestationFingerprint(""))
	// Output:
	// ed25519:abc1234…d4f2
	// --
}

func ExampleRegionList() {
	fmt.Println(RegionList([]string{"iad1", "ord1", "sea1", "dfw1"}, 1))
	fmt.Println(RegionList([]string{"iad1", "ord1"}, 3))
	fmt.Println(RegionList(nil, 3))
	// Output:
	// iad1, +3 more
	// iad1, ord1
	// --
}

func ExampleToolchainSpec() {
	fmt.Println(ToolchainSpec(map[string]string{"java": "17", "node": "20"}))
	fmt.Println(ToolchainSpec(nil))
	// Output:
	// java=17, node=20
	// --
}

func ExampleHumanLabel() {
	type billingModel string
	labels := map[billingModel]string{
		"wall-clock": "Wall-clock time",
		"active-cpu": "Active CPU only",
	}
	fmt.Println(HumanLabel(billingModel("wall-clock"), labels))
	fmt.Println(HumanLabel(billingModel("invocation"), labels))
	// Output:
	// Wall-clock time
	// --
}
