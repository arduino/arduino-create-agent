import os

def running_on_ci():
    """
    Returns whether the program is running on a CI environment
    """
    return 'GITHUB_WORKFLOW' in os.environ
