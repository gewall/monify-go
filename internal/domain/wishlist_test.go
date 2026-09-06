package domain

import "testing"

func TestEstimateWishlist(t *testing.T) {
	items := []WishlistItem{
		{Name: "Headphone", TargetAmountMinor: 300000_00, SavedAmountMinor: 100000_00}, // rem 200k
		{Name: "Laptop", TargetAmountMinor: 1500000_00, SavedAmountMinor: 0},           // rem 1.5jt
		{Name: "Done", TargetAmountMinor: 50000_00, SavedAmountMinor: 50000_00},        // funded
	}
	const surplus = 100000_00 // 100k/bulan

	par := EstimateWishlist(items, surplus, false)
	if par[0].MonthsNeeded != 2 {
		t.Errorf("parallel headphone months = %d, want 2", par[0].MonthsNeeded)
	}
	if par[1].MonthsNeeded != 15 {
		t.Errorf("parallel laptop months = %d, want 15", par[1].MonthsNeeded)
	}
	if !par[2].AlreadyFunded {
		t.Errorf("item 3 should be already funded")
	}

	seq := EstimateWishlist(items, surplus, true)
	if seq[0].MonthsNeeded != 2 {
		t.Errorf("seq headphone months = %d, want 2", seq[0].MonthsNeeded)
	}
	if seq[1].MonthsNeeded != 17 { // (200k + 1.5jt) / 100k = 17
		t.Errorf("seq laptop months = %d, want 17", seq[1].MonthsNeeded)
	}

	neg := EstimateWishlist(items, -5000, false)
	if neg[0].Estimable {
		t.Errorf("negative surplus should not be estimable")
	}
}
