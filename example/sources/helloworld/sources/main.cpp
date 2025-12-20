#include <iostream>

#include <hellohelper/hellohelper.hpp>

int main()
{
    std::cout << "Hello, " << hellohelper::get_world() << "!" << std::endl;
    return 0;
}