package htm

import (
	//"fmt"
	"github.com/skelterjohn/go.matrix"
	//"github.com/stretchr/testify/assert"
	"github.com/zacg/testify/assert"
	//"math/big"
	//"github.com/stretchr/testify/mock"
	"math"
	"testing"
)

func getConnected(perm []float64, sp *SpatialPooler) (int, []bool) {
	numcon := 0
	connected := make([]bool, len(perm))
	for i := 0; i < len(perm); i++ {
		if perm[i] >= sp.SynPermConnected {
			numcon++
			connected[i] = true
		} else {
			connected[i] = false
		}
	}

	return numcon, connected
}

func TestPermanenceInit(t *testing.T) {
	sp := SpatialPooler{}
	sp.InputDimensions = []int{1, 10}
	sp.numInputs = 10
	sp.SynPermConnected = 0.1
	sp.SynPermActiveInc = 0.1
	//sp.raisePermanenceToThreshold = Mock()

	sp.PotentialRadius = 2
	connectedPct := 1.0
	mask := []bool{true, true, true, false, false, false, false, false, true, true}
	perm := sp.initPermanence(mask, connectedPct)
	numcon, connected := getConnected(perm, &sp)

	if numcon != 5 {
		t.Errorf("numcon was %v expected 5", numcon)
	}
	maxThresh := sp.SynPermConnected + sp.SynPermActiveInc/4

	for i := 0; i < len(perm); i++ {
		if perm[i] > maxThresh {
			t.Errorf("perm %v was %v higher than threshold", i, perm[i])
		}
	}

	connectedPct = 0
	numcon = 0
	perm = sp.initPermanence(mask, connectedPct)
	numcon, connected = getConnected(perm, &sp)
	if numcon != 0 {
		t.Errorf("numcon was %v expected 0", numcon)
	}

	if len(connected) != 5 {
		return
	}

	connectedPct = 0.5
	sp.PotentialRadius = 100
	sp.numInputs = 100
	mask = make([]bool, 100)
	for i := 0; i < len(mask); i++ {
		mask[i] = true
	}

	perm = sp.initPermanence(mask, connectedPct)
	numcon, connected = getConnected(perm, &sp)

	if !(numcon > 0) {
		t.Errorf("numcon was %v expected greater than 0", numcon)
	}

	if numcon >= sp.numInputs {
		t.Errorf("numcon was %v expected less than inputs count", numcon)
	}

	minThresh := sp.SynPermActiveInc / 2.0
	connThresh := sp.SynPermConnected

	for i := 0; i < len(perm); i++ {
		if perm[i] < minThresh {
			t.Errorf("perm %v was %v less than min threshold", i, perm[i])
		}
		if perm[i] >= connThresh {
			t.Errorf("perm %v was %v not less than connection threshold", i, perm[i])
		}
	}

}

func TestRaisePermanenceThreshold(t *testing.T) {

	sp := SpatialPooler{}
	sp.InputDimensions = []int{1, 5}
	sp.ColumnDimensions = []int{1, 5}
	sp.numColumns = 5
	sp.numInputs = 5
	sp.SynPermConnected = 0.1
	sp.StimulusThreshold = 3
	sp.SynPermBelowStimulusInc = 0.01
	sp.SynPermMin = 0
	sp.SynPermMax = 1

	elms := make(map[int]float64, 25)
	sp.permanences = matrix.MakeSparseMatrix(elms, 5, 5)

	p := [][]float64{
		{0.0, 0.11, 0.095, 0.092, 0.01},
		{0.12, 0.15, 0.02, 0.12, 0.09},
		{0.51, 0.081, 0.025, 0.089, 0.31},
		{0.18, 0.0601, 0.11, 0.011, 0.03},
		{0.011, 0.011, 0.011, 0.011, 0.011},
	}
	AddDenseToSparseHelper(p, sp.permanences)

	sp.connectedSynapses = NewSparseBinaryMatrixFromDense([][]bool{
		{false, true, false, false, false},
		{true, true, false, true, false},
		{true, false, false, false, true},
		{true, false, true, false, false},
		{false, false, false, false, false},
	})
	sp.connectedCounts = []int{1, 3, 2, 2, 0}

	truePermanences := [][]float64{
		{0.01, 0.12, 0.105, 0.102, 0.02},
		{0.12, 0.15, 0.02, 0.12, 0.09},
		{0.53, 0.101, 0.045, 0.109, 0.33},
		{0.22, 0.1001, 0.15, 0.051, 0.07},
		{0.101, 0.101, 0.101, 0.101, 0.101},
	}

	maskPP := []int{0, 1, 2, 3, 4}

	for i := 0; i < sp.numColumns; i++ {
		perm := SparseMatrixToArray(sp.permanences.GetRowVector(i))
		sp.raisePermanenceToThreshold(perm, maskPP)
		for j := 0; j < sp.numInputs; j++ {
			//if truePermanences[i][j] != perm[j] {
			if !AlmostEqual(truePermanences[i][j], perm[j]) {
				t.Errorf("truePermances: %v != perm: %v", truePermanences[i][j], perm[j])
			}
		}
	}

}

func TestStripNever(t *testing.T) {
	sp := SpatialPooler{}

	sp.activeDutyCycles = []float64{0.5, 0.1, 0, 0.2, 0.4, 0}
	activeColumns := []int{0, 1, 2, 4}
	stripped := sp.stripNeverLearned(activeColumns)
	trueStripped := []int{0, 1, 4}
	t.Logf("stripped", stripped)
	for i := 0; i < len(trueStripped); i++ {
		if stripped[i] != trueStripped[i] {
			t.Errorf("stripped %v was %v expected %v", i, stripped[i], trueStripped[i])
		}
	}

	sp.activeDutyCycles = []float64{0.9, 0, 0, 0, 0.4, 0.3}
	activeColumns = []int{0, 1, 2, 3, 4, 5}
	stripped = sp.stripNeverLearned(activeColumns)
	trueStripped = []int{0, 4, 5}
	for i := 0; i < len(trueStripped); i++ {
		if stripped[i] != trueStripped[i] {
			t.Errorf("stripped %v was %v expected %v", i, stripped[i], trueStripped[i])
		}
	}

	sp.activeDutyCycles = []float64{0, 0, 0, 0, 0, 0}
	activeColumns = []int{0, 1, 2, 3, 4, 5}
	stripped = sp.stripNeverLearned(activeColumns)
	if len(stripped) != 0 {
		t.Errorf("Expected empty stripped was %v", stripped)
	}

	sp.activeDutyCycles = []float64{1, 1, 1, 1, 1, 1}
	activeColumns = []int{0, 1, 2, 3, 4, 5}
	stripped = sp.stripNeverLearned(activeColumns)
	trueStripped = []int{0, 1, 2, 3, 4, 5}
	for i := 0; i < len(trueStripped); i++ {
		if stripped[i] != trueStripped[i] {
			t.Errorf("stripped %v was %v expected %v", i, stripped[i], trueStripped[i])
		}
	}

}

func TestAvgConnectedSpanForColumnND(t *testing.T) {
	sp := SpatialPooler{}
	sp.InputDimensions = []int{4, 4, 2, 5}
	sp.numInputs = ProdInt(sp.InputDimensions)
	sp.numColumns = 5
	sp.ColumnDimensions = []int{0, 1, 2, 3, 4}

	sp.connectedSynapses = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)

	connected := make([]bool, sp.numInputs)
	connected[(1*40)+(0*10)+(1*5)+(0*1)] = true
	connected[(1*40)+(0*10)+(1*5)+(1*1)] = true
	connected[(3*40)+(2*10)+(1*5)+(0*1)] = true
	connected[(3*40)+(0*10)+(1*5)+(0*1)] = true
	connected[(1*40)+(0*10)+(1*5)+(3*1)] = true
	connected[(2*40)+(2*10)+(1*5)+(0*1)] = true

	//# span: 3 3 1 4, avg = 11/4
	sp.connectedSynapses.ReplaceRow(0, connected)

	connected2 := make([]bool, sp.numInputs)
	connected2[(2*40)+(0*10)+(1*5)+(0*1)] = true
	connected2[(2*40)+(0*10)+(0*5)+(0*1)] = true
	connected2[(3*40)+(0*10)+(0*5)+(0*1)] = true
	connected2[(3*40)+(0*10)+(1*5)+(0*1)] = true
	//spn: 2 1 2 1, avg = 6/4
	sp.connectedSynapses.ReplaceRow(1, connected2)

	connected3 := make([]bool, sp.numInputs)
	connected3[(0*40)+(0*10)+(1*5)+(4*1)] = true
	connected3[(0*40)+(0*10)+(0*5)+(3*1)] = true
	connected3[(0*40)+(0*10)+(0*5)+(1*1)] = true
	connected3[(1*40)+(0*10)+(0*5)+(2*1)] = true
	connected3[(0*40)+(0*10)+(1*5)+(1*1)] = true
	connected3[(3*40)+(3*10)+(1*5)+(1*1)] = true
	// span: 4 4 2 4, avg = 14/4
	sp.connectedSynapses.ReplaceRow(2, connected3)

	connected4 := make([]bool, sp.numInputs)
	connected4[(3*40)+(3*10)+(1*5)+(4*1)] = true
	connected4[(0*40)+(0*10)+(0*5)+(0*1)] = true

	// span: 4 4 2 5, avg = 15/4
	sp.connectedSynapses.ReplaceRow(3, connected4)

	connected5 := make([]bool, sp.numInputs)
	//# span: 0 0 0 0, avg = 0
	sp.connectedSynapses.ReplaceRow(4, connected5)

	//t.Logf("width: %v", sp.connectedSynapses.Width)

	trueAvgConnectedSpan := []float64{11.0 / 4.0, 6.0 / 4.0, 14.0 / 4.0, 15.0 / 4.0, 0.0}

	for i, tspan := range trueAvgConnectedSpan {
		connectedSpan := sp.avgConnectedSpanForColumnND(i)
		if connectedSpan != tspan {
			t.Errorf("Connected span was: %v expected: %v", connectedSpan, tspan)
		}
	}

}

func TestAvgColumnsPerInput(t *testing.T) {
	sp := SpatialPooler{}
	sp.ColumnDimensions = []int{2, 2, 2, 2}
	sp.InputDimensions = []int{4, 4, 4, 4}

	if sp.avgColumnsPerInput() != 0.5 {
		t.Errorf("Expected %v avg columns, was: %v", 0.5, sp.avgColumnsPerInput())
	}

	sp.ColumnDimensions = []int{2, 2, 2, 2}
	sp.InputDimensions = []int{7, 5, 1, 3}
	//2/7 0.4 2 0.666
	trueAvgColumnPerInput := (2.0/7 + 2.0/5 + 2.0/1 + 2/3.0) / 4
	if sp.avgColumnsPerInput() != trueAvgColumnPerInput {
		t.Errorf("Expected %v avg columns, was: %v", trueAvgColumnPerInput, sp.avgColumnsPerInput())
	}

	sp.ColumnDimensions = []int{3, 3}
	sp.InputDimensions = []int{3, 3}
	// 1 1
	trueAvgColumnPerInput = 1
	if sp.avgColumnsPerInput() != trueAvgColumnPerInput {
		t.Errorf("Expected %v avg columns, was: %v", trueAvgColumnPerInput, sp.avgColumnsPerInput())
	}

	sp.ColumnDimensions = []int{25}
	sp.InputDimensions = []int{5}
	// 5
	trueAvgColumnPerInput = 5
	if sp.avgColumnsPerInput() != trueAvgColumnPerInput {
		t.Errorf("Expected %v avg columns, was: %v", trueAvgColumnPerInput, sp.avgColumnsPerInput())
	}

	sp.ColumnDimensions = []int{3, 3, 3, 5, 5, 6, 6}
	sp.InputDimensions = []int{3, 3, 3, 5, 5, 6, 6}
	//1 1 1 1 1 1 1
	trueAvgColumnPerInput = 1
	if sp.avgColumnsPerInput() != trueAvgColumnPerInput {
		t.Errorf("Expected %v avg columns, was: %v", trueAvgColumnPerInput, sp.avgColumnsPerInput())
	}

	sp.ColumnDimensions = []int{3, 6, 9, 12}
	sp.InputDimensions = []int{3, 3, 3, 3}
	// 1 2 3 4
	trueAvgColumnPerInput = 2.5
	if sp.avgColumnsPerInput() != trueAvgColumnPerInput {
		t.Errorf("Expected %v avg columns, was: %v", trueAvgColumnPerInput, sp.avgColumnsPerInput())
	}

}

/*
func TestUpdateInhibitionRadius(t *testing.T) {
	sp := spMock{}

	// Test global inhibition case
	sp.GlobalInhibition = true
	sp.ColumnDimensions = []int{57, 31, 2}
	sp.numColumns = 3
	sp.updateInhibitionRadius()
	expected := 57
	assert.Equal(t, expected, sp.inhibitionRadius)

	sp.GlobalInhibition = false

	//sp.avgConnectedSpanForColumnND = Mock(return_value = 3)
	//sp.AvgColumnsPerInput = Mock(return_value = 4)
	trueInhibitionRadius := 6
	//((3 * 4) - 1) / 2 => round up
	sp.updateInhibitionRadius()
	assert.Equal(t, trueInhibitionRadius, sp.inhibitionRadius)

	//Test clipping at 1.0
	sp.GlobalInhibition = false
	//sp.AvgConnectedSpanForColumnND = Mock(return_value = 0.5)
	//sp.AvgColumnsPerInput = Mock(return_value = 1.2)
	trueInhibitionRadius = 1
	sp.updateInhibitionRadius()
	assert.Equal(t, trueInhibitionRadius, sp.inhibitionRadius)

	//Test rounding up
	sp.GlobalInhibition = false
	//sp.AvgConnectedSpanForColumnND = Mock(return_value = 2.4)
	//sp.AvgColumnsPerInput = Mock(return_value = 2)
	trueInhibitionRadius = 2
	// ((2 * 2.4) - 1) / 2.0 => round up
	sp.updateInhibitionRadius()
	assert.Equal(t, trueInhibitionRadius, sp.inhibitionRadius)

}
*/

func TestCalculateOverlap(t *testing.T) {
	sp := SpatialPooler{}
	sp.numInputs = 10
	sp.numColumns = 5

	sp.connectedSynapses = NewSparseBinaryMatrixFromDense([][]bool{
		{true, true, true, true, true, true, true, true, true, true},
		{false, false, true, true, true, true, true, true, true, true},
		{false, false, false, false, true, true, true, true, true, true},
		{false, false, false, false, false, false, true, true, true, true},
		{false, false, false, false, false, false, false, false, true, true},
	})
	t.Log(sp.connectedSynapses.ToString())
	sp.connectedCounts = []int{10, 8, 6, 4, 2}
	inputVector := make([]bool, sp.numInputs)
	overlaps := sp.calculateOverlap(inputVector)
	overlapsPct := sp.calculateOverlapPct(overlaps)
	trueOverlaps := []int{0, 0, 0, 0, 0}
	trueOverlapsPct := []float64{0, 0, 0, 0, 0}
	t.Logf("pct", overlapsPct)
	assert.Equal(t, trueOverlaps, overlaps)
	assert.Equal(t, overlapsPct, trueOverlapsPct)

	inputVector = make([]bool, sp.numInputs)
	for i := 0; i < len(inputVector); i++ {
		inputVector[i] = true
	}
	overlaps = sp.calculateOverlap(inputVector)
	overlapsPct = sp.calculateOverlapPct(overlaps)
	trueOverlaps = []int{10, 8, 6, 4, 2}
	trueOverlapsPct = []float64{1, 1, 1, 1, 1}
	t.Logf("pct", overlapsPct)
	assert.Equal(t, trueOverlaps, overlaps)
	assert.Equal(t, overlapsPct, trueOverlapsPct)

	inputVector = make([]bool, sp.numInputs)
	inputVector[9] = true
	t.Logf("input", inputVector)
	overlaps = sp.calculateOverlap(inputVector)
	overlapsPct = sp.calculateOverlapPct(overlaps)
	trueOverlaps = []int{1, 1, 1, 1, 1}
	trueOverlapsPct = []float64{0.1, 0.125, 1.0 / 6, 0.25, 0.5}
	t.Logf("pct", overlapsPct)
	assert.Equal(t, trueOverlaps, overlaps)
	assert.Equal(t, trueOverlapsPct, overlapsPct)

	//Zig-zag
	sp.connectedSynapses = NewSparseBinaryMatrixFromDense([][]bool{
		{true, false, false, false, false, true, false, false, false, false},
		{false, true, false, false, false, false, true, false, false, false},
		{false, false, true, false, false, false, false, true, false, false},
		{false, false, false, true, false, false, false, false, true, false},
		{false, false, false, false, true, false, false, false, false, true},
	})
	sp.connectedCounts = []int{2.0, 2.0, 2.0, 2.0, 2.0}
	inputVector = make([]bool, sp.numInputs)
	inputVector[0] = true
	inputVector[2] = true
	inputVector[4] = true
	inputVector[6] = true
	inputVector[8] = true
	t.Logf("input", inputVector)
	overlaps = sp.calculateOverlap(inputVector)
	overlapsPct = sp.calculateOverlapPct(overlaps)
	trueOverlaps = []int{1, 1, 1, 1, 1}
	trueOverlapsPct = []float64{0.5, 0.5, 0.5, 0.5, 0.5}
	t.Logf("pct", overlapsPct)
	assert.Equal(t, trueOverlaps, overlaps)
	assert.Equal(t, trueOverlapsPct, overlapsPct)

}

/*
sp = self._sp
    density = 0.3
    sp._numColumns = 10
    overlaps = numpy.array([1, 2, 1, 4, 8, 3, 12, 5, 4, 1])
    active = list(sp._inhibitColumnsGlobal(overlaps, density))
    trueActive = numpy.zeros(sp._numColumns)
    trueActive = [4, 6, 7]
    self.assertListEqual(list(trueActive), active)

    density = 0.5
    sp._numColumns = 10
    overlaps = numpy.array(range(10))
    active = list(sp._inhibitColumnsGlobal(overlaps, density))
    trueActive = numpy.zeros(sp._numColumns)
    trueActive = range(5, 10)
    self.assertListEqual(trueActive, active)



*/

func TestInhibitColumnsGlobal(t *testing.T) {
	sp := SpatialPooler{}
	density := 0.3
	sp.numColumns = 10
	overlaps := []float64{1, 2, 1, 4, 8, 3, 12, 5, 4, 1}
	active := sp.inhibitColumnsGlobal(overlaps, density)
	trueActive := []int{4, 6, 7}
	assert.Equal(t, trueActive, active)

	density = 0.5
	overlaps = []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	active = sp.inhibitColumnsGlobal(overlaps, density)
	trueActive = []int{5, 6, 7, 8, 9}
	assert.Equal(t, trueActive, active)

}

func TestGetNeighborsND(t *testing.T) {
	sp := SpatialPooler{}

	dimensions := []int{5, 7, 2}
	var layout [5][7][2]int

	counter := 0
	for i := range layout {
		for j := range layout[i] {
			for k := range layout[i][j] {
				layout[i][j][k] = counter
				counter++
			}
		}
	}

	radius := 1
	x := 1
	y := 3
	z := 2

	columnIndex := layout[z][y][x]

	neighbors := sp.getNeighborsND(columnIndex, dimensions, radius, true)

	var expected []int

	for i := radius * -1; i <= radius; i++ {
		for j := radius * -1; j <= radius; j++ {
			for k := radius * -1; k <= radius; k++ {
				zprime := (z + i) % dimensions[0]
				yprime := (y + j) % dimensions[1]
				xprime := (x + k) % dimensions[2]
				if layout[zprime][yprime][xprime] != columnIndex && !ContainsInt(layout[zprime][yprime][xprime], expected) {
					expected = append(expected, layout[zprime][yprime][xprime])
				}

			}
		}
	}

	assert.Equal(t, expected, neighbors)

	dimensions = []int{5, 7, 9}
	var layoutb [5][7][9]int
	counter = 0
	for i := range layoutb {
		for j := range layoutb[i] {
			for k := range layoutb[i][j] {
				layoutb[i][j][k] = counter
				counter++
			}
		}
	}

	radius = 3
	x = 0
	y = 0
	z = 3
	columnIndex = layoutb[z][y][x]
	neighbors = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	expected = []int{}
	for i := radius * -1; i <= radius; i++ {
		for j := radius * -1; j <= radius; j++ {
			for k := radius * -1; k <= radius; k++ {
				zprime := Mod((z + i), (dimensions[0]))
				yprime := Mod((y + j), (dimensions[1]))
				xprime := Mod((x + k), (dimensions[2]))

				if layoutb[zprime][yprime][xprime] != columnIndex && !ContainsInt(layoutb[zprime][yprime][xprime], expected) {
					expected = append(expected, layoutb[zprime][yprime][xprime])
				}

			}
		}
	}

	assert.Equal(t, expected, neighbors)

	dimensions = []int{5, 10, 7, 6}
	var layoutc [5][10][7][6]int
	counter = 0
	for i := range layoutc {
		for j := range layoutc[i] {
			for k := range layoutc[i][j] {
				for m := range layoutc[i][j][k] {
					layoutc[i][j][k][m] = counter
					counter++
				}
			}
		}
	}

	radius = 4
	w := 2
	x = 5
	y = 6
	z = 2
	columnIndex = layoutc[z][y][x][w]
	neighbors = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	expected = []int{}
	for i := radius * -1; i <= radius; i++ {
		for j := radius * -1; j <= radius; j++ {
			for k := radius * -1; k <= radius; k++ {
				for m := radius * -1; m <= radius; m++ {
					zprime := Mod((z + i), (dimensions[0]))
					yprime := Mod((y + j), (dimensions[1]))
					xprime := Mod((x + k), (dimensions[2]))
					wprime := Mod((w + m), (dimensions[3]))

					if layoutc[zprime][yprime][xprime][wprime] != columnIndex && !ContainsInt(layoutc[zprime][yprime][xprime][wprime], expected) {
						expected = append(expected, layoutc[zprime][yprime][xprime][wprime])
					}

				}

			}
		}
	}

	assert.Equal(t, expected, neighbors)

	layoutd := []bool{false, false, true, false, true, false, false, false}
	columnIndex = 3
	dimensions = []int{8}
	radius = 1
	mask := sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)

	for idx, val := range layoutd {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	layoute := []bool{false, true, true, false, true, true, false, false}
	columnIndex = 3
	dimensions = []int{8}
	radius = 2

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)

	for idx, val := range layoute {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	// Wrap around
	layoutf := []bool{false, true, true, false, false, false, true, true}
	columnIndex = 0
	dimensions = []int{8}
	radius = 2

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)

	for idx, val := range layoutf {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	//Radius too big
	layoutg := []bool{true, true, true, true, true, true, false, true}
	columnIndex = 6
	dimensions = []int{8}
	radius = 20

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)

	for idx, val := range layoutg {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	//These are all the same tests from 2D
	ints := [][]int{{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 1, 1, 1, 0},
		{0, 1, 0, 1, 0},
		{0, 1, 1, 1, 0},
		{0, 0, 0, 0, 0}}

	layouth := NewSparseBinaryMatrixFromInts(ints)
	t.Log(layouth.ToString())

	columnIndex = 3*5 + 2
	dimensions = []int{6, 5}
	radius = 1

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)
	t.Log("1d", layouth.Flatten())
	for idx, val := range layouth.Flatten() {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	ints = [][]int{{0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 0, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1}}

	layoutj := NewSparseBinaryMatrixFromInts(ints)
	t.Log(layouth.ToString())

	columnIndex = 3*5 + 2
	radius = 2

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)
	t.Log("1d", layouth.Flatten())
	for idx, val := range layoutj.Flatten() {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	//Radius too big
	ints = [][]int{{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 0, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1}}

	layoutk := NewSparseBinaryMatrixFromInts(ints)
	t.Log(layouth.ToString())

	columnIndex = 3*5 + 2
	radius = 7

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)
	t.Log("1d", layoutk.Flatten())
	for idx, val := range layoutk.Flatten() {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	//Wrap-around
	ints = [][]int{{1, 0, 0, 1, 1},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{1, 0, 0, 1, 1},
		{1, 0, 0, 1, 0}}

	layoutl := NewSparseBinaryMatrixFromInts(ints)
	t.Log(layoutl.ToString())

	columnIndex = 29
	radius = 1

	mask = sp.getNeighborsND(columnIndex, dimensions, radius, true)
	t.Log("mask", mask)
	t.Log("1d", layoutl.Flatten())
	for idx, val := range layoutl.Flatten() {
		if ContainsInt(idx, mask) {
			assert.Equal(t, true, val)
		} else {
			assert.Equal(t, false, val)
		}
	}

	//No wrap around
	columnIndex = 8
	radius = 2
	dimensions = []int{10}
	neighbors = sp.getNeighborsND(columnIndex, dimensions, radius, false)
	expected = []int{6, 7, 9}
	assert.Equal(t, expected, neighbors)

}

func TestInhibitColumnsLocal(t *testing.T) {
	sp := SpatialPooler{}
	density := 0.5
	sp.numColumns = 10
	sp.ColumnDimensions = []int{sp.numColumns}
	sp.inhibitionRadius = 2
	overlaps := []float64{1, 2, 7, 0, 3, 4, 16, 1, 1.5, 1.7}
	// L W  W  L  L  W   W  L   L    W
	trueActive := []int{1, 2, 5, 6, 9}
	active := sp.inhibitColumnsLocal(overlaps, density)
	assert.Equal(t, trueActive, active)

	//Test add to winners
	density = 0.3333
	sp.inhibitionRadius = 3
	overlaps = []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	// W W L L W W L L L W
	trueActive = []int{0, 1, 4, 5, 8}
	active = sp.inhibitColumnsLocal(overlaps, density)
	assert.Equal(t, trueActive, active)

}

func UpdateBoostFactorsTest(t *testing.T) {
	sp := SpatialPooler{}
	sp.MaxBoost = 10.0
	sp.numColumns = 6
	sp.minActiveDutyCycles = make([]float64, sp.numColumns)
	for i, _ := range sp.minActiveDutyCycles {
		sp.minActiveDutyCycles[i] = -0.0000001
	}
	sp.activeDutyCycles = []float64{0.1, 0.3, 0.02, 0.04, 0.7, 0.12}
	sp.boostFactors = make([]float64, sp.numColumns)
	trueBoostFactors := []float64{1, 1, 1, 1, 1, 1}
	sp.updateBoostFactors()

	for i, _ := range sp.boostFactors {
		assert.Equal(t, trueBoostFactors[i], sp.boostFactors[i])
	}

	sp.MaxBoost = 10.0
	sp.numColumns = 6
	sp.minActiveDutyCycles = []float64{0.1, 0.3, 0.02, 0.04, 0.7, 0.12}
	sp.activeDutyCycles = []float64{0.1, 0.3, 0.02, 0.04, 0.7, 0.12}
	trueBoostFactors = []float64{1, 1, 1, 1, 1, 1}
	sp.updateBoostFactors()

	for i, _ := range sp.boostFactors {
		diff := math.Abs(trueBoostFactors[i] - sp.boostFactors[i])
		assert.True(t, diff <= 0.0000001)
	}

	sp.MaxBoost = 10.0
	sp.numColumns = 6
	sp.minActiveDutyCycles = []float64{0.1, 0.2, 0.02, 0.03, 0.7, 0.12}
	sp.activeDutyCycles = []float64{0.01, 0.02, 0.002, 0.003, 0.07, 0.012}
	trueBoostFactors = []float64{9.1, 9.1, 9.1, 9.1, 9.1, 9.1}
	sp.updateBoostFactors()
	for i, _ := range sp.boostFactors {
		diff := math.Abs(trueBoostFactors[i] - sp.boostFactors[i])
		assert.True(t, diff <= 0.0000001)
	}

	sp.MaxBoost = 10.0
	sp.numColumns = 6
	sp.minActiveDutyCycles = []float64{0.1, 0.2, 0.02, 0.03, 0.7, 0.12}
	sp.activeDutyCycles = make([]float64, sp.numColumns)
	trueBoostFactors = MakeSliceFloat64(6, sp.MaxBoost)
	sp.updateBoostFactors()
	for i, _ := range sp.boostFactors {
		diff := math.Abs(trueBoostFactors[i] - sp.boostFactors[i])
		assert.True(t, diff <= 0.0000001)
	}

}

func TestUpdateDutyCycleHelper(t *testing.T) {

	dc := []float64{1000.0, 1000.0, 1000.0, 1000.0, 1000.0}
	period := 1000
	newvals := make([]int, 5)
	actual := updateDutyCyclesHelper(dc, newvals, period)
	expected := []float64{999, 999, 999, 999, 999}
	assert.Equal(t, expected, actual)

	FillSliceInt(newvals, 1000)
	actual = updateDutyCyclesHelper(dc, newvals, period)
	assert.Equal(t, dc, actual)

	newvals = []int{2000, 4000, 5000, 6000, 7000}
	expected = []float64{1001, 1003, 1004, 1005, 1006}
	actual = updateDutyCyclesHelper(dc, newvals, period)
	assert.Equal(t, expected, actual)

	dc = []float64{1000, 800, 600, 400, 2000}
	FillSliceInt(newvals, 0)
	period = 2
	actual = updateDutyCyclesHelper(dc, newvals, period)
	expected = []float64{500, 400, 300, 200, 1000}
	assert.Equal(t, expected, actual)
}

func TestUpdatePermanencesForColumn(t *testing.T) {

	sp := SpatialPooler{}
	sp.InputDimensions = []int{5}
	sp.ColumnDimensions = []int{5}
	sp.numColumns = 5
	sp.numInputs = 5
	sp.SynPermConnected = 0.1
	sp.SynPermTrimThreshold = 0.05
	sp.connectedCounts = make([]int, sp.numColumns)
	elms := make(map[int]float64, 25)
	sp.permanences = matrix.MakeSparseMatrix(elms, sp.numColumns, sp.numInputs)
	sp.potentialPools = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)
	sp.connectedSynapses = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)
	sp.SynPermMax = 1
	sp.SynPermMin = 0

	permanences := [][]float64{
		{-0.10, 0.500, 0.400, 0.010, 0.020},
		{0.300, 0.010, 0.020, 0.120, 0.090},
		{0.070, 0.050, 1.030, 0.190, 0.060},
		{0.180, 0.090, 0.110, 0.010, 0.030},
		{0.200, 0.101, 0.050, -0.09, 1.100}}

	/*
	   These are the 'true permanences' reflected in trueConnectedSynapses
	   truePermanences = SparseMatrix(
	   [[0.000, 0.500, 0.400, 0.000, 0.000],
	   Clip - - Trim Trim
	   [0.300, 0.000, 0.000, 0.120, 0.090],
	   - Trim Trim - -
	   [0.070, 0.050, 1.000, 0.190, 0.060],
	   - - Clip - -
	   [0.180, 0.090, 0.110, 0.000, 0.000],
	   - - - Trim Trim
	   [0.200, 0.101, 0.050, 0.000, 1.000]])
	   - - - Clip Clip
	*/

	trueConnectedSynapses := Make2DBool([][]int{{0, 1, 1, 0, 0},
		{1, 0, 0, 1, 0},
		{0, 0, 1, 1, 0},
		{1, 0, 1, 0, 0},
		{1, 1, 0, 0, 1}})

	trueConnectedCounts := []int{2, 2, 2, 2, 3}

	for i := 0; i < sp.numColumns; i++ {
		sp.updatePermanencesForColumn(permanences[i], i, true)
		assert.Equal(t, trueConnectedSynapses[i], sp.connectedSynapses.GetDenseRow(i))
	}

	assert.Equal(t, trueConnectedCounts, sp.connectedCounts)

}

func TestAdaptSynapses(t *testing.T) {

	sp := SpatialPooler{}
	sp.InputDimensions = []int{8}
	sp.ColumnDimensions = []int{4}
	sp.numInputs = 8
	sp.numColumns = 4
	sp.SynPermInactiveDec = 0.01
	sp.SynPermActiveInc = 0.1
	sp.SynPermTrimThreshold = 0.05
	sp.connectedSynapses = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)
	sp.connectedCounts = make([]int, sp.numColumns)
	sp.SynPermMax = 1
	sp.SynPermMin = 0

	ints := [][]int{{1, 1, 1, 1, 0, 0, 0, 0},
		{1, 0, 0, 0, 1, 1, 0, 1},
		{0, 0, 1, 0, 0, 0, 1, 0},
		{1, 0, 0, 0, 0, 0, 1, 0}}
	sp.potentialPools = NewSparseBinaryMatrixFromInts(ints)

	inputVector := []bool{true, false, false, true, true, false, true, false}
	activeColumns := []int{0, 1, 2}

	floats := []float64{0.200, 0.120, 0.090, 0.040, 0.000, 0.000, 0.000, 0.000,
		0.150, 0.000, 0.000, 0.000, 0.180, 0.120, 0.000, 0.450,
		0.000, 0.000, 0.014, 0.000, 0.000, 0.000, 0.110, 0.000,
		0.040, 0.000, 0.000, 0.000, 0.000, 0.000, 0.178, 0.000}
	elms := make(map[int]float64, len(floats))
	for i, val := range floats {
		elms[i] = val
	}
	sp.permanences = matrix.MakeSparseMatrix(elms, 4, 8)

	truePermanences := [][]float64{
		{0.300, 0.110, 0.080, 0.140, 0.000, 0.000, 0.000, 0.000},
		// Inc Dec Dec Inc - - - -
		{0.250, 0.000, 0.000, 0.000, 0.280, 0.110, 0.000, 0.440},
		// Inc - - - Inc Dec - Dec
		{0.000, 0.000, 0.000, 0.000, 0.000, 0.000, 0.210, 0.000},
		// - - Trim - - - Inc -
		{0.040, 0.000, 0.000, 0.000, 0.000, 0.000, 0.178, 0.000}}
	// - - - - - - - -

	sp.adaptSynapses(inputVector, activeColumns)

	for i := 0; i < sp.numColumns; i++ {
		for j := 0; j < sp.numInputs; j++ {
			expected := truePermanences[i][j]
			actual := sp.permanences.Get(i, j)
			assert.AlmostEqual(t, expected, actual)
		}
	}

	ints = [][]int{{1, 1, 1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 0, 0, 0, 0},
		{0, 0, 1, 1, 1, 0, 0, 0},
		{1, 0, 0, 0, 0, 0, 1, 0}}

	sp.potentialPools = NewSparseBinaryMatrixFromInts(ints)

	inputVector = []bool{true, false, false, true, true, false, true, false}
	activeColumns = []int{0, 1, 2}

	floats = []float64{0.200, 0.120, 0.090, 0.000, 0.000, 0.000, 0.000, 0.000,
		0.000, 0.017, 0.232, 0.400, 0.000, 0.000, 0.000, 0.000,
		0.000, 0.000, 0.014, 0.051, 0.730, 0.000, 0.000, 0.000,
		0.170, 0.000, 0.000, 0.000, 0.000, 0.000, 0.380, 0.000}
	for i, val := range floats {
		elms[i] = val
	}
	sp.permanences = matrix.MakeSparseMatrix(elms, 4, 8)

	truePermanences = [][]float64{
		{0.30, 0.110, 0.080, 0.000, 0.000, 0.000, 0.000, 0.000},
		// Inc Dec Dec - - - - -
		{0.000, 0.000, 0.222, 0.500, 0.000, 0.000, 0.000, 0.000},
		// - Trim Dec Inc - - - -
		{0.000, 0.000, 0.000, 0.151, 0.830, 0.000, 0.000, 0.000},
		// - - Trim Inc Inc - - -
		{0.170, 0.000, 0.000, 0.000, 0.000, 0.000, 0.380, 0.000}}
	// - - - - - - - -

	for i := 0; i < sp.numColumns; i++ {
		for j := 0; j < sp.numInputs; j++ {
			expected := truePermanences[i][j]
			actual := sp.permanences.Get(i, j)
			assert.AlmostEqual(t, expected, actual)
		}
	}

}

func TestIsUpdateRound(t *testing.T) {
	sp := SpatialPooler{}
	sp.UpdatePeriod = 50

	sp.IterationNum = 1
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 39
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 50
	assert.Equal(t, true, sp.isUpdateRound())
	sp.IterationNum = 1009
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 1250
	assert.Equal(t, true, sp.isUpdateRound())

	sp.UpdatePeriod = 125
	sp.IterationNum = 0
	assert.Equal(t, true, sp.isUpdateRound())
	sp.IterationNum = 200
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 249
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 1330
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 1249
	assert.Equal(t, false, sp.isUpdateRound())
	sp.IterationNum = 1375
	assert.Equal(t, true, sp.isUpdateRound())
}

func TestUpdateMinDutyCyclesGlobal(t *testing.T) {
	sp := SpatialPooler{}
	sp.MinPctActiveDutyCycles = 0.02
	sp.MinPctOverlapDutyCycles = 0.01
	sp.numColumns = 5
	sp.overlapDutyCycles = []float64{0.06, 1, 3, 6, 0.5}
	sp.activeDutyCycles = []float64{0.6, 0.07, 0.5, 0.4, 0.3}
	sp.minOverlapDutyCycles = make([]float64, sp.numColumns)
	sp.minActiveDutyCycles = make([]float64, sp.numColumns)
	sp.updateMinDutyCyclesGlobal()
	trueMinActiveDutyCycles := MakeSliceFloat64(sp.numColumns, 0.02*0.6)
	trueMinOverlapDutyCycles := MakeSliceFloat64(sp.numColumns, 0.01*6)

	assert.Equal(t, 5, len(sp.minActiveDutyCycles))
	assert.Equal(t, 5, len(sp.minOverlapDutyCycles))
	for i := 0; i < sp.numColumns; i++ {
		assert.AlmostEqual(t, trueMinActiveDutyCycles[i], sp.minActiveDutyCycles[i])
		assert.AlmostEqual(t, trueMinOverlapDutyCycles[i], sp.minOverlapDutyCycles[i])
	}

	sp.MinPctOverlapDutyCycles = 0.015
	sp.MinPctActiveDutyCycles = 0.03
	sp.numColumns = 5
	sp.overlapDutyCycles = []float64{0.86, 2.4, 0.03, 1.6, 1.5}
	sp.activeDutyCycles = []float64{0.16, 0.007, 0.15, 0.54, 0.13}
	sp.updateMinDutyCyclesGlobal()
	trueMinOverlapDutyCycles = MakeSliceFloat64(sp.numColumns, 0.015*2.4)
	for i := 0; i < sp.numColumns; i++ {
		assert.AlmostEqual(t, trueMinOverlapDutyCycles[i], sp.minOverlapDutyCycles[i])
	}

	sp.MinPctOverlapDutyCycles = 0.015
	sp.MinPctActiveDutyCycles = 0.03
	sp.numColumns = 5
	sp.overlapDutyCycles = MakeSliceFloat64(sp.numColumns, 0)
	sp.activeDutyCycles = MakeSliceFloat64(sp.numColumns, 0)
	sp.updateMinDutyCyclesGlobal()
	trueMinActiveDutyCycles = MakeSliceFloat64(sp.numColumns, 0)
	trueMinOverlapDutyCycles = MakeSliceFloat64(sp.numColumns, 0)

	assert.Equal(t, 5, len(sp.minActiveDutyCycles))
	assert.Equal(t, 5, len(sp.minOverlapDutyCycles))
	for i := 0; i < sp.numColumns; i++ {
		assert.AlmostEqual(t, trueMinActiveDutyCycles[i], sp.minActiveDutyCycles[i])
		assert.AlmostEqual(t, trueMinOverlapDutyCycles[i], sp.minOverlapDutyCycles[i])
	}

}

// func TestUpdateMinDutyCyclesLocal(t *testing.T) {
//TODO: implement
// }

func TestBumpUpWeakColumns(t *testing.T) {
	sp := SpatialPooler{}
	sp.numInputs = 8
	sp.InputDimensions = []int{8}
	sp.numColumns = 5
	sp.ColumnDimensions = []int{5}
	sp.SynPermBelowStimulusInc = 0.01
	sp.SynPermTrimThreshold = 0.05
	sp.overlapDutyCycles = []float64{0, 0.009, 0.1, 0.001, 0.002}
	sp.minOverlapDutyCycles = MakeSliceFloat64(5, 5*0.01)
	sp.SynPermInactiveDec = 0.01
	sp.SynPermActiveInc = 0.1
	sp.connectedSynapses = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)
	sp.connectedCounts = make([]int, sp.numColumns)
	sp.SynPermMax = 1
	sp.SynPermMin = 0

	//sp._potentialPools = SparseBinaryMatrix(
	ints := [][]int{{1, 1, 1, 1, 0, 0, 0, 0},
		{1, 0, 0, 0, 1, 1, 0, 1},
		{0, 0, 1, 0, 1, 1, 1, 0},
		{1, 1, 1, 0, 0, 0, 1, 0},
		{1, 1, 1, 1, 1, 1, 1, 1}}
	sp.potentialPools = NewSparseBinaryMatrixFromInts(ints)

	floats := []float64{0.200, 0.120, 0.090, 0.040, 0.000, 0.000, 0.000, 0.000,
		0.150, 0.000, 0.000, 0.000, 0.180, 0.120, 0.000, 0.450,
		0.000, 0.000, 0.014, 0.000, 0.032, 0.044, 0.110, 0.000,
		0.041, 0.000, 0.000, 0.000, 0.000, 0.000, 0.178, 0.000,
		0.100, 0.738, 0.045, 0.002, 0.050, 0.008, 0.208, 0.034}
	elms := make(map[int]float64, len(floats))
	for i, val := range floats {
		elms[i] = val
	}
	sp.permanences = matrix.MakeSparseMatrix(elms, 5, 8)

	truePermanences := [][]float64{
		{0.210, 0.130, 0.100, 0.000, 0.000, 0.000, 0.000, 0.000},
		// Inc Inc Inc Trim - - - -
		{0.160, 0.000, 0.000, 0.000, 0.190, 0.130, 0.000, 0.460},
		// Inc - - - Inc Inc - Inc
		{0.000, 0.000, 0.014, 0.000, 0.032, 0.044, 0.110, 0.000}, //unchanged
		// - - - - - - - -
		{0.051, 0.000, 0.000, 0.000, 0.000, 0.000, 0.188, 0.000},
		// Inc Trim Trim - - - Inc -
		{0.110, 0.748, 0.055, 0.000, 0.060, 0.000, 0.218, 0.000}}

	sp.bumpUpWeakColumns()

	for i := 0; i < sp.numColumns; i++ {
		for j := 0; j < sp.numInputs; j++ {
			assert.AlmostEqual(t, truePermanences[i][j], sp.permanences.Get(i, j))
		}
	}

}

//----- Helper functions -------------

func AlmostEqual(a, b float64) bool {
	ar := RoundPrec(a, 2)
	br := RoundPrec(b, 2)
	return ar == br
}

func SparseMatrixToArray(m *matrix.SparseMatrix) []float64 {
	result := make([]float64, m.Cols())
	for i := 0; i < m.Cols(); i++ {
		result[i] = m.Get(0, i)
	}
	return result
}

func AddDenseToSparseHelper(dense [][]float64, m *matrix.SparseMatrix) {
	for r := 0; r < len(dense); r++ {
		for c := 0; c < len(dense[r]); c++ {
			m.Set(r, c, dense[r][c])
		}
	}
}
