package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

//===============================================
// Rule 41 Substrings and memory leaks
//===============================================

type account struct {
	ID      int
	Balance int
}

func concepts() {
	/// Less error-prone, since we don't have to handle conditional expressions, iteration variables, etc.
	s := []string{"a", "b", "c"}
	for i, v := range s {
		fmt.Printf("index=%d, value=%s", i, v)
	}

	accounts := []account{
		{
			ID:      1,
			Balance: 100,
		},
		{
			ID:      2,
			Balance: 200,
		},
		{
			ID:      3,
			Balance: 300,
		},
	}

	/// This result is in [{100}, {200}, {300}]
	/// since iterator is based on copy of the value.
	for _, account := range accounts {
		account.Balance += 1000 // Does not update the original value!
	}

	s1 := "Hello world!"
	// s2 := s1[:5] // Hello  -> This causes memory leak since s2 is a view of s1.
	s2 := string([]rune(s1)[:5]) // Hello -> This is correct.
	fmt.Println(s2)

}

//===============================================
// Rule 42 Not knowing which type of receiver to use
//===============================================

type bankCustomer struct {
	balance float32
}

// This works
func (c *bankCustomer) add(operation float32) {
	c.balance += operation
}

// But this does not work!
func (c bankCustomer) misusedAdd(operation float32) {
	c.balance += operation
}

// / 1. use a pointer of receiver is a large object
// / 2. use a value to enforce immutability, or receivcer is a map, function or channel
// / 3. prefer value if type is small array or struct, or baisc types such as int, float... etc
// / Mixing receiver types should be avoided, but there's a counterexample (time.Time)
func test() {
	bankCustomer := bankCustomer{
		balance: 100,
	}

	bankCustomer.add(10)
	fmt.Println(bankCustomer.balance)

	customer := Customer{Age: 25, Name: "Minho"}
	// Results in "cusotmer is invalid : <nil>"
	// err is always not nil here. since it's "error" object with "nil" value inside.
	if err := customer.Validate(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
}

//===============================================
// Rule 43 Never using named result parameters
//===============================================

// This function default-initializes b, and automatically returns it.
// Specifying name to result can be helpful for code readability.
// But, should be used in suitable cases, eg) interfaces, or cases where it's not obvious what the result is.
// Also, it can be used to shorten the implementation, by default-initializing the result.
func f(a int) (b int) {
	b = a + 1
	return
}

//===============================================
// Rule 44 Unintended side deffects with named result parameters
//===============================================

type loc struct {
	lat float32
}

func (l loc) validateAddress(address string) bool { // Method
	return true
}

func (l loc) getCoordinates(ctx context.Context, address string) (lat, lng float32, err error) {
	isValid := l.validateAddress(address)
	if !isValid {
		return 0, 0, fmt.Errorf("invalid address")
	}

	if ctx.Err() != nil {
		// This would have been a compile error if we didn't use named result parameters.
		// However, in this case, it returns "nil" silently without warning the programmer, which might be unintended.
		return 0, 0, err // Caution! err is always "nil"!
	}

	// Another option is to use naked return statement,
	// but it's not recommended since it breaks rule "we shouldn't mix naked returns and returns with arguments"
	if err = ctx.Err(); err != nil {
		return
	}

	if address == "" {
		return 0, 0, fmt.Errorf("address is empty")
	}
	return l.lat, 0, nil
}

//===============================================
// Rule 45 Returning a nil receiver
//===============================================

type MultiError struct {
	errs []string
}

func (m *MultiError) myAdd(err error) {
	m.errs = append(m.errs, err.Error())
}

func (m *MultiError) Error() string {
	return strings.Join(m.errs, ";")
}

type Customer struct {
	Age  int
	Name string
}

func (c Customer) Validate() error {
	var m *MultiError
	if c.Age < 0 {
		m = &MultiError{}
		m.myAdd(errors.New("age is negative"))
	}
	if c.Name == "" {
		if m == nil {
			m = &MultiError{}
		}
		m.myAdd(errors.New("name is nil"))
	}

	// fix : check for nil before wrapping it with error interface
	// if m != nil{
	// 	return m
	// }
	// return nil

	// If all checks are valid, argument provided to return isn't nil, but it will be an error interface pointing to nil.
	return m
}

//===============================================
// Rule 46 Using a filename as a function input
//===============================================

// / If we write in this way, we have no abastraction over io.Reader interface.
// / So, if we want to accept from another source, we have to make separate functions for each source.
// / If we were to test this function, we have to make separate files for each case.
func countEmptyLinesInFile(filename string) (emptyLines int, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == "" {
			emptyLines++
		}
	}

	return emptyLines, nil
}

// / this would be much better solution since code is resuable, and we can test it with different sources.
func countEmptyLinesInFile2(r io.Reader) (emptyLines int, err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if scanner.Text() == "" {
			emptyLines++
		}
	}

	return emptyLines, scanner.Err()
}

//===============================================
// Rule 47 Ignoring how defer arguments and receivers are evaluated
//===============================================

func deferTest() error {
	i := 0
	j := 0
	defer func(i int) {
		fmt.Println(i, j) // i is evaluated when "defer" is evaluated, while j is captured as reference, and is evaluated when this function is evaluated
	}(i) // To avoid such behavior, we can pass i as a reference as &i
	i++
	j++

	return nil
}

type DeferStruct struct {
	id string
}

func (d DeferStruct) print() {
	fmt.Println(d.id)
}

func methodDeferTest() {
	s := DeferStruct{id: "123"}
	defer s.print()

	s.id = "456"
	// This will print "123" because the value of s is captured when defer is evaluated, not when the function is evaluated.
	// So, if we want to capture the value of s, we can pass it as a reference as &s.
	// But, this is not a good practice, because it's not clear what the value of s is.
	// So, we should avoid using defer with a large object.
}

func methodDeferTest2() {
	s := &DeferStruct{id: "123"}
	defer s.print()

	s.id = "456"
	// This will print "456" since s is a pointer.
}

//===============================================
// Rule 48 panicking
//===============================================

func panicTest() {
	// This will print "Recovered from panic: your program is doomed"
	// Deferred function is always called regardless of panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	fmt.Println("before panic")
	panic("your program is doomed")
	fmt.Println("after panic") // This is unreachable.
}

/// Use panics in
/// 1. Code status is invalid
/// 2. Dependencies failed to initialize

//===============================================
// Rule 49 Ignoring when to wrap an error
//===============================================

// We wrap errors for
// 1. Add additional information to the error
// 2. Mark an error as a specific error

type BarError struct {
	Err error
}

func (b BarError) Error() string {
	return "bar failed: " + b.Err.Error()
}

func failingFunction() error {
	err := errors.New("I'm sorry, but You're fucked up bro")
	if err != nil {
		// Wrap the error with BarError.
		return BarError{Err: err}
	}
	return err
}

func failingFunction2() error {
	err := errors.New("You're really doomed now")
	if err != nil {
		/// This creats implicit wrapper (fmt.wrapError) around the error, just like what we did with BarError.
		return fmt.Errorf("failed: %w", err)
	}
	return err
}

func failingFunction3() error {
	err := errors.New("Are you out of your mind?")
	if err != nil {
		/// This "transforms" the error, doesn't wrap the error. Therefore, this is more restrictive in some sense, but it gives continuity from the perspective of the caller.
		// Source error is no longer aviailable to the caller. This decouples the caller from the source error.
		return fmt.Errorf("failed: %v", err)
	}
	return err
}

//===============================================
// Rule 50 Checking an error type inaccurately
//===============================================

type transientError struct {
	err error
}

func (t transientError) Error() string {
	return fmt.Sprintf("traisient error: %v", t.err)
}

func getTransactionAmountFromDB(transactionID string) (float32, error) {
	return 0, errors.New("transaction not found")
}

func getTransactionAmount(transactionID string) (float32, error) {
	// This function returns different error types depending on the situation.
	if len(transactionID) != 5 {
		// Returns fmt.Errorf if identifier is invalid
		return 0, fmt.Errorf("id is invalid: %s", transactionID)
	}

	amount, err := getTransactionAmountFromDB(transactionID)
	if err != nil {
		// Returns trasientError if DB fails.
		return 0, transientError{err: err}
	}

	return amount, nil
}

func getTransactionAmountFromDB2(transactionID string) (float32, error) {
	return 0, transientError{err: errors.New("transaction not found")}
}

func getTransactionAmount2(transactionID string) (float32, error) {
	// This function returns different error types depending on the situation.
	if len(transactionID) != 5 {
		// Returns fmt.Errorf if identifier is invalid
		return 0, fmt.Errorf("id is invalid: %s", transactionID)
	}

	amount, err := getTransactionAmountFromDB(transactionID)
	if err != nil {
		// Returns trasientError if DB fails.
		return 0, fmt.Errorf("transient error: %w", err)
	}

	return amount, nil
}

// Here's the handler that chekcs the error type.
func handler(w http.ResponseWriter, r *http.Request) {
	// transactionID := r.URL.Query().Get("transaction")

	// Now, returned error is not directly transientError, but it's wrapped with fmt.Errorf.
	// In this case, it will be always 400 error.
	_, err := getTransactionAmount2(r.URL.Query().Get("id"))
	if err != nil {
		// This is a problem since it's not checking the error type accurately.
		// If the error is a transientError, it will be handled as a normal error.
		// This is a problem since it's not checking the error type accurately.
		switch err := err.(type) {
		case transientError:
			// Return 503 error
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			// Return 400 error
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

// Here's the handler that chekcs the error type.
func fixedHandler(w http.ResponseWriter, r *http.Request) {
	// transactionID := r.URL.Query().Get("transaction")

	// Now, returned error is not directly transientError, but it's wrapped with fmt.Errorf.
	// In this case, it will be always 400 error.
	_, err := getTransactionAmount2(r.URL.Query().Get("id"))
	if err != nil {
		// Now, errors.As recursively unwraps the error, and checks if it's a transientError.
		// This feature is added from Go 1.13.
		if errors.As(err, &transientError{}) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
}

//===============================================
// Rule 51 Check error value inaccurately
//===============================================

var sentinelError = errors.New("sentinel error")

// Use wrapped error.
func query_w(sentinel bool) error {
	if sentinel {
		return fmt.Errorf("sentinel error: %w", sentinelError)
	}
	return nil
}

// Use raw error.
func query(sentinel bool) error {
	if sentinel {
		return sentinelError
	}
	return nil
}

// Use == to check if the error is a specific error with wrapped error.
func checkErrorValue() {
	err := query_w(true)

	if err != nil {
		if err == sentinelError {
			fmt.Println("sentinel error")
		} else {
			fmt.Println("unknown error")
		}
	}

}

// Use == to check if the error is a specific error with raw error.
func checkErrorValue2() {
	err := query(true)

	if err != nil {
		if err == sentinelError {
			fmt.Println("sentinel error")
		} else {
			fmt.Println("unknown error")
		}
	}
}

// Use errors.Is to check if the error is a specific error with wrapped error.
func checkErrorValue3() {
	err := query_w(true)

	if err != nil {
		if errors.Is(err, sentinelError) {
			fmt.Println("sentinel error")
		} else {
			fmt.Println("unknown error")
		}
	}
}

// ===============================================
// Rule 52. 에러를 두 번 처리하지 마라
// ===============================================

type Route struct {
	Distance float32
	Duration float32
}

func validateCoordinates(lat, lng float32) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("invalid latitude: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("invalid longitude: %f", lng)
	}
	return nil
}

func GetRoute(srcLat, srcLng, dstLat, dstLng float32) (Route, error) {
	err := validateCoordinates(srcLat, srcLng)
	if err != nil {
		return Route{}, fmt.Errorf("failed to validate source coordinates: %w", err)
	}

	err = validateCoordinates(dstLat, dstLng)
	if err != nil {
		return Route{}, fmt.Errorf("failed to validate destination coordinates: %w", err)
	}

	return Route{
		Distance: 10,
		Duration: 10,
	}, nil
}
