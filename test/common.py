import os

def running_on_ci():
    """
    Returns whether the program is running on a CI environment
    """
    val = os.getenv("GITHUB_WORKFLOW")
    return val is not None
