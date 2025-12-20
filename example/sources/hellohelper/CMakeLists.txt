cmake_minimum_required(VERSION 3.10)

project(HelloHelper LANGUAGES CXX)

include(GNUInstallDirs)

add_library(hello_helper
        sources/hellohelper.cpp
        headers/hellohelper/hellohelper.hpp)

add_library(HelloHelper::hello_helper ALIAS hello_helper)

target_include_directories(hello_helper
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/headers>
        $<INSTALL_INTERFACE:include>
)

set_target_properties(hello_helper PROPERTIES EXPORT_NAME hello_helper)

install(
    DIRECTORY headers/
    DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}
)

install(
    TARGETS hello_helper
    EXPORT hellohelper_exports
)

export(
    EXPORT hellohelper_exports
    FILE hellohelper-config.cmake
    NAMESPACE HelloHelper::
)

install(
    EXPORT hellohelper_exports
    FILE hellohelper-config.cmake
    NAMESPACE HelloHelper::
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/hellohelper
)
