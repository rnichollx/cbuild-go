cmake_minimum_required(VERSION 3.10)

project(HelloWorld)

find_package(HelloHelper REQUIRED)

add_executable(helloworld sources/main.cpp)

target_link_libraries(helloworld PRIVATE HelloHelper::hello_helper)