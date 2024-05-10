const char *getDefaultBrowserName();

const char *installCert(const char *path);
const char *uninstallCert();
const bool certInKeychain();

const char *getExpirationDate(long *expirationDate);