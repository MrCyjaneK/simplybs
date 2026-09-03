# Python 3.13 removed the stdlib pipes module. mx 7.4 / Graal 22 still imports it.
from shlex import quote

__all__ = ["quote"]
