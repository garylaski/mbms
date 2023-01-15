CXX = g++
CXXFLAGS = -O3
LIB = -ltag -lz -lsqlite3 -lcurl -lxml2
INC = -I/usr/include/taglib -I/usr/include/libxml2
all:  generate_db.cpp
	$(CXX) $(CXXFLAGS) $(LIB) $(INC) generate_db.cpp -o generate_db
clean:
	$(RM) generate_db ../mb.db
