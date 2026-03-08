NAME := ft_turing
SRC := ft_turing.py

.PHONY: all clean fclean re check test

all: check
	chmod +x $(NAME)

check:
	@command -v python3 >/dev/null 2>&1 || (echo "Missing tool: python3" && exit 1)
	python3 -m py_compile $(SRC)

clean:
	rm -rf __pycache__

fclean: clean

re: fclean all

test: all
	python3 -m unittest discover -s tests -p "test_*.py" -v
