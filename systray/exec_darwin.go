package systray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

char **makeCharArray(int size) {
	return calloc(sizeof(char*), size);
}

void setCharArray(char **a, int n, char *s) {
	a[n] = s;
}

void freeCharArray(char **a, int size) {
	int i;
	for (i = 0; i < size; i++) {
		free(a[i]);
	}
	free(a);
}

void runApplication(const char *path, const char **argv, int argc) {
	NSMutableArray<NSString *> *stringArray = [NSMutableArray array];
	for (int i=0; i<argc; i++) {
		NSString *arg = [NSString stringWithCString:argv[i] encoding:NSUTF8StringEncoding];
		[stringArray addObject:arg];
	}
	NSArray<NSString *> *arguments = [NSArray arrayWithArray:stringArray];

	NSWorkspace *ws = [NSWorkspace sharedWorkspace];
	NSURL *url = [NSURL fileURLWithPath:@(path) isDirectory:NO];

	NSWorkspaceOpenConfiguration* configuration = [NSWorkspaceOpenConfiguration new];
	//[configuration setEnvironment:env];
	[configuration setPromptsUserIfNeeded:YES];
	[configuration setCreatesNewApplicationInstance:YES];
	[configuration setArguments:arguments];
	dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
	[ws openApplicationAtURL:url configuration:configuration completionHandler:^(NSRunningApplication* app, NSError* error) {
		dispatch_semaphore_signal(semaphore);
	}];
	dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);
}
*/
import "C"
import (
	"os/exec"
	"path/filepath"
	"unsafe"

	"github.com/sirupsen/logrus"
)

func execApp(path string, args ...string) error {
	if filepath.Ext(path) != ".app" {
		// If not .app, fallback to standard process execution
		logrus.WithField("path", path).WithField("args", args).Info("Running new app with os/exec.Exec")
		cmd := exec.Command(path, args...)
		return cmd.Start()
	}

	logrus.WithField("path", path).WithField("args", args).Info("Running new app with openApplicationAtURL")
	argc := C.int(len(args))
	argv := C.makeCharArray(argc)
	for i, arg := range args {
		C.setCharArray(argv, C.int(i), C.CString(arg))
	}

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.runApplication(cpath, argv, argc)

	C.freeCharArray(argv, argc)
	return nil
}
