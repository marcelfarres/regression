package regression

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/skelterjohn/go.matrix"
)

var (
	errNotEnoughData = errors.New("Not enough data points")
	errTooManyvars   = errors.New("Not enough observations to to support this many variables.")
)

type Regression struct {
	names             describe
	data              []*dataPoint
	coeff             map[int]float64
	R2                float64
	Varianceobserved  float64
	VariancePredicted float64
	initialised       bool
	Formula           string
}

type dataPoint struct {
	Observed  float64
	Variables []float64
	Predicted float64
	Error     float64
}

type describe struct {
	obs  string
	vars map[int]string
}

// Creates a new dataPoint
func DataPoint(obs float64, vars []float64) *dataPoint {
	return &dataPoint{Observed: obs, Variables: vars}
}

// Predict updates the "Predicted" value for the input dataPoint
func (r *Regression) Predict(vars []float64) (float64, error) {
	if !r.initialised {
		return 0, errNotEnoughData
	}
	p := r.Coeff(0)
	for j := 1; j < len(r.data[0].Variables)+1; j++ {
		p += r.Coeff(j) * vars[j-1]
	}
	return p, nil
}

// Set the name of the observed value
func (r *Regression) SetObserved(name string) {
	r.names.obs = name
}

// Get the name of the observed value
func (r *Regression) GetObserved() string {
	return r.names.obs
}

// Set the name of variable i
func (r *Regression) SetVar(i int, name string) {
	if len(r.names.vars) == 0 {
		r.names.vars = make(map[int]string, 5)
	}
	r.names.vars[i] = name
}

// Get the name of variable i
func (r *Regression) GetVar(i int) string {
	x := r.names.vars[i]
	if x == "" {
		s := []string{"X", strconv.Itoa(i)}
		return strings.Join(s, "")
	}
	return x
}

// Train the regression with some data points
func (r *Regression) Train(d ...*dataPoint) {
	r.data = append(r.data, d...)
	if len(r.data) > 2 {
		r.initialised = true
	}
}

// Run the regression
func (r *Regression) Run() error {
	if !r.initialised {
		return errNotEnoughData
	}

	observations := len(r.data)
	numOfvars := len(r.data[0].Variables)

	if observations < (numOfvars + 1) {
		return errTooManyvars
	}

	// Create some blank variable space
	observed := matrix.Zeros(observations, 1)
	variables := matrix.Zeros(observations, numOfvars+1)

	for i := 0; i < observations; i++ {
		observed.Set(i, 0, r.data[i].Observed)
		for j := 0; j < numOfvars+1; j++ {
			if j == 0 {
				variables.Set(i, 0, 1)
			} else {
				variables.Set(i, j, r.data[i].Variables[j-1])
			}
		}
	}

	// Now run the regression
	n := variables.Cols()
	q, reg := variables.QR()
	qty, err := q.Transpose().Times(observed)
	if err != nil {
		return err
	}
	c := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		c[i] = qty.Get(i, 0)
		for j := i + 1; j < n; j++ {
			c[i] -= c[j] * reg.Get(i, j)
		}
		c[i] /= reg.Get(i, i)
	}

	// Output the regression results
	r.coeff = make(map[int]float64, numOfvars)
	for i, val := range c {
		r.coeff[i] = val
		if i == 0 {
			r.Formula = fmt.Sprintf("Predicted = %.2f", val)
		} else {
			r.Formula += fmt.Sprintf(" + %v*%.2f", r.GetVar(i-1), val)
		}
	}

	r.calcPredicted()
	r.calcVariance()
	r.calcR2()
	return nil
}

// Return the calulated coefficient for variable i
func (r *Regression) Coeff(i int) float64 {
	if len(r.coeff) == 0 {
		return 0
	}
	return r.coeff[i]
}

func (r *Regression) calcPredicted() string {
	observations := len(r.data)
	var predicted float64
	var output string
	for i := 0; i < observations; i++ {
		r.data[i].Predicted, _ = r.Predict(r.data[i].Variables)
		r.data[i].Error = r.data[i].Predicted - r.data[i].Observed

		output += fmt.Sprintf("%v. observed = %v, Predicted = %v, Error = %v", i, r.data[i].Observed, predicted, r.data[i].Error)
	}
	return output
}

func (r *Regression) calcVariance() string {
	observations := len(r.data)
	var obtotal, prtotal, obvar, prvar float64
	for i := 0; i < observations; i++ {
		obtotal += r.data[i].Observed
		prtotal += r.data[i].Predicted
	}
	obaverage := obtotal / float64(observations)
	praverage := prtotal / float64(observations)

	for i := 0; i < observations; i++ {
		obvar += math.Pow(r.data[i].Observed-obaverage, 2)
		prvar += math.Pow(r.data[i].Predicted-praverage, 2)
	}
	r.Varianceobserved = obvar / float64(observations)
	r.VariancePredicted = prvar / float64(observations)
	return fmt.Sprintf("N = %v\nVariance observed = %v\nVariance Predicted = %v\n", observations, r.Varianceobserved, r.VariancePredicted)
}

func (r *Regression) calcR2() string {
	r.R2 = r.VariancePredicted / r.Varianceobserved
	return fmt.Sprintf("R2 = %.2f", r.R2)
}

func (r *Regression) calcResiduals() string {
	str := fmt.Sprintf("Residuals:\nobserved|\tPredicted|\tResidual\n")
	for _, d := range r.data {
		str += fmt.Sprintf("%.2f|\t%.2f|\t%.2f\n", d.Observed, d.Predicted, d.Observed-d.Predicted)
	}
	str += "\n"
	return str
}

// Display a dataPoint as a string
func (d *dataPoint) String() string {
	str := fmt.Sprintf("%.2f", d.Observed)
	for _, v := range d.Variables {
		str += fmt.Sprintf("|\t%.2f", v)
	}
	return str
}

// Display a regression as a string
func (r *Regression) String() string {
	if !r.initialised {
		return errNotEnoughData.Error()
	}
	str := fmt.Sprintf("%v", r.GetObserved())
	for i := 0; i < len(r.names.vars); i++ {
		str += fmt.Sprintf("|\t%v", r.GetVar(i))
	}
	str += "\n"
	for _, d := range r.data {
		str += fmt.Sprintf("%v\n", d)
	}
	fmt.Println(r.calcResiduals())
	str += fmt.Sprintf("\nN = %v\nVariance observed = %v\nVariance Predicted = %v", len(r.data), r.Varianceobserved, r.VariancePredicted)
	str += fmt.Sprintf("\nR2 = %v\n", r.R2)
	return str
}
