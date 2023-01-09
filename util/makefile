CXX = g++
CXXFLAGS = -g
LIB = -ltag -lz -lsqlite3
INC = -I/usr/include/taglib
all: generate_db.cpp
	$(CXX) $(CXXFLAGS) $(LIB) $(INC) generate_db.cpp -o generate_db
clean:
	$(RM) main
