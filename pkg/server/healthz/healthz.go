package healthz

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/server/httplog"
	"k8s.io/klog/v2"
)

func AddToContainer(container *restful.Container, path string, checks ...HealthChecker) error {
	name := strings.Split(strings.TrimPrefix(path, "/"), "/")[0]
	container.Handle(path, HandleRootHealth(name, nil, checks...))

	for _, check := range checks {
		container.Handle(fmt.Sprintf("%s/%v", path, check.Name()), adaptCheckToHandler(check))
	}

	return nil
}

func Handler(container *restful.Container, checks ...HealthChecker) error {
	if len(checks) == 0 {
		klog.V(4).Info("No default health checks specified. Installing the ping handler.")
		checks = []HealthChecker{PingHealthz}
	}
	return AddToContainer(container, "/healthz", checks...)
}

type HealthChecker interface {
	Name() string
	Check(req *http.Request) error
}

func adaptCheckToHandler(checks HealthChecker) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Header().Set("X-Content-Type-Options", "nosniff")

		err := checks.Check(request)
		if err != nil {
			http.Error(writer, fmt.Sprintf("internal server error: %v", err), http.StatusInternalServerError)
		} else {
			fmt.Fprint(writer, "ok")
		}
	}
}

func HandleRootHealth(name string, firstTimeHealthy func(), checks ...HealthChecker) http.HandlerFunc {
	var notifyOnce sync.Once
	return func(w http.ResponseWriter, r *http.Request) {
		excluded := getExcludedChecks(r)

		var failedVerboseLogOutput bytes.Buffer
		var failedChecks []string
		var individualCheckOutput bytes.Buffer
		for _, check := range checks {
			if excluded.Has(check.Name()) {
				excluded.Delete(check.Name())
				fmt.Fprintf(&individualCheckOutput, "[+]%s exclude: ok\n", check.Name())
				continue
			}
			if err := check.Check(r); err != nil {
				fmt.Fprintf(&individualCheckOutput, "[-]%s failed: reason with\n", check.Name())
				fmt.Fprintf(&failedVerboseLogOutput, "[-]%s failed: %v\n", check.Name(), err)
				failedChecks = append(failedChecks, check.Name())
			} else {
				fmt.Fprintf(&individualCheckOutput, "[+]%s ok\n", check.Name())
			}
		}

		if excluded.Len() > 0 {
			fmt.Fprintf(&individualCheckOutput, "warn: some health checks can't be excluded: no matches for %s\n", formatQuoted(excluded.UnsortedList()...))
			klog.Warningf("can't exclude some health checks, no health checks are installed matching %s", formatQuoted(excluded.UnsortedList()...))
		}

		if len(failedChecks) > 0 {
			klog.V(2).Infof("%s check failed: %s\n%v", strings.Join(failedChecks, ","), name, failedVerboseLogOutput.String())
			httplog.SetStacktracePredicate(r.Context(), func(int) bool { return false })
			http.Error(w, fmt.Sprintf("%s%s check failed", individualCheckOutput.String(), name), http.StatusInternalServerError)
			return
		}

		if firstTimeHealthy != nil {
			notifyOnce.Do(firstTimeHealthy)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, found := r.URL.Query()["verbose"]; !found {
			fmt.Fprint(w, "ok")
			return
		}

		individualCheckOutput.WriteTo(w)
		fmt.Fprintf(w, "%s check healthz passed\n", name)
	}

}

func getExcludedChecks(r *http.Request) sets.Set[string] {
	checks, found := r.URL.Query()["exclude"]
	if found {
		return sets.New(checks...)
	}
	return sets.New[string]()
}

func formatQuoted(names ...string) string {
	q := make([]string, 0, len(names))
	for _, name := range names {
		q = append(q, fmt.Sprintf("%q", name))
	}
	return strings.Join(q, ",")
}

var PingHealthz HealthChecker = ping{}

type ping struct{}

func (ping) Name() string {
	return "ping"
}

func (ping) Check(_ *http.Request) error {
	return nil
}
