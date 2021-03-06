package nghttp2

import (
	"github.com/bradfitz/http2/hpack"
	"github.com/tatsuhiro-t/spdy"
	"net/http"
	"testing"
)

// TestS3H1PlainGET tests whether simple SPDY GET request works.
func TestS3H1PlainGET(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, noopHandler)
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1PlainGET",
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}

	want := 200
	if got := res.status; got != want {
		t.Errorf("status = %v; want %v", got, want)
	}
}

// TestS3H1BadRequestCL tests that server rejects request whose
// content-length header field value does not match its request body
// size.
func TestS3H1BadRequestCL(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, noopHandler)
	defer st.Close()

	// we set content-length: 1024, but the actual request body is
	// 3 bytes.
	res, err := st.spdy(requestParam{
		name:   "TestS3H1BadRequestCL",
		method: "POST",
		header: []hpack.HeaderField{
			pair("content-length", "1024"),
		},
		body: []byte("foo"),
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}

	want := spdy.ProtocolError
	if got := res.spdyRstErrCode; got != want {
		t.Errorf("res.spdyRstErrCode = %v; want %v", got, want)
	}
}

// TestS3H1MultipleRequestCL tests that server rejects request with
// multiple Content-Length request header fields.
func TestS3H1MultipleRequestCL(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not forward bad request")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1MultipleRequestCL",
		header: []hpack.HeaderField{
			pair("content-length", "1"),
			pair("content-length", "1"),
		},
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	want := 400
	if got := res.status; got != want {
		t.Errorf("status: %v; want %v", got, want)
	}
}

// TestS3H1InvalidRequestCL tests that server rejects request with
// Content-Length which cannot be parsed as a number.
func TestS3H1InvalidRequestCL(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not forward bad request")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1InvalidRequestCL",
		header: []hpack.HeaderField{
			pair("content-length", ""),
		},
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	want := 400
	if got := res.status; got != want {
		t.Errorf("status: %v; want %v", got, want)
	}
}

// TestS3H1GenerateVia tests that server generates Via header field to and
// from backend server.
func TestS3H1GenerateVia(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Via"), "1.1 nghttpx"; got != want {
			t.Errorf("Via: %v; want %v", got, want)
		}
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1GenerateVia",
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	if got, want := res.header.Get("Via"), "1.1 nghttpx"; got != want {
		t.Errorf("Via: %v; want %v", got, want)
	}
}

// TestS3H1AppendVia tests that server adds value to existing Via
// header field to and from backend server.
func TestS3H1AppendVia(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Via"), "foo, 1.1 nghttpx"; got != want {
			t.Errorf("Via: %v; want %v", got, want)
		}
		w.Header().Add("Via", "bar")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1AppendVia",
		header: []hpack.HeaderField{
			pair("via", "foo"),
		},
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	if got, want := res.header.Get("Via"), "bar, 1.1 nghttpx"; got != want {
		t.Errorf("Via: %v; want %v", got, want)
	}
}

// TestS3H1NoVia tests that server does not add value to existing Via
// header field to and from backend server.
func TestS3H1NoVia(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1", "--no-via"}, t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Via"), "foo"; got != want {
			t.Errorf("Via: %v; want %v", got, want)
		}
		w.Header().Add("Via", "bar")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1NoVia",
		header: []hpack.HeaderField{
			pair("via", "foo"),
		},
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	if got, want := res.header.Get("Via"), "bar"; got != want {
		t.Errorf("Via: %v; want %v", got, want)
	}
}

// TestS3H1HeaderFieldBuffer tests that request with header fields
// larger than configured buffer size is rejected.
func TestS3H1HeaderFieldBuffer(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1", "--header-field-buffer=10"}, t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("execution path should not be here")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1HeaderFieldBuffer",
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	if got, want := res.spdyRstErrCode, spdy.InternalError; got != want {
		t.Errorf("res.spdyRstErrCode: %v; want %v", got, want)
	}
}

// TestS3H1HeaderFields tests that request with header fields more
// than configured number is rejected.
func TestS3H1HeaderFields(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1", "--max-header-fields=1"}, t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("execution path should not be here")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H1HeaderFields",
		// we have at least 5 pseudo-header fields sent, and
		// that ensures that buffer limit exceeds.
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	if got, want := res.spdyRstErrCode, spdy.InternalError; got != want {
		t.Errorf("res.spdyRstErrCode: %v; want %v", got, want)
	}
}

// TestS3H1InvalidMethod tests that server rejects invalid method with
// 501.
func TestS3H1InvalidMethod(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1"}, t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not forward this request")
	})
	defer st.Close()

	res, err := st.spdy(requestParam{
		name:   "TestS3H1InvalidMethod",
		method: "get",
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	if got, want := res.status, 501; got != want {
		t.Errorf("status: %v; want %v", got, want)
	}
}

// TestS3H2ConnectFailure tests that server handles the situation that
// connection attempt to HTTP/2 backend failed.
func TestS3H2ConnectFailure(t *testing.T) {
	st := newServerTesterTLS([]string{"--npn-list=spdy/3.1", "--http2-bridge"}, t, noopHandler)
	defer st.Close()

	// simulate backend connect attempt failure
	st.ts.Close()

	res, err := st.spdy(requestParam{
		name: "TestS3H2ConnectFailure",
	})
	if err != nil {
		t.Fatalf("Error st.spdy() = %v", err)
	}
	want := 503
	if got := res.status; got != want {
		t.Errorf("status: %v; want %v", got, want)
	}
}
