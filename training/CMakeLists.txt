cmake_minimum_required(VERSION 3.20)
project(kingdoms_sim CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Release)
endif()

include(FetchContent)
FetchContent_Declare(
  json
  GIT_REPOSITORY https://github.com/nlohmann/json.git
  GIT_TAG v3.11.3
)
FetchContent_MakeAvailable(json)

file(GLOB_RECURSE KINGDOMS_SIM_SOURCES CONFIGURE_DEPENDS src/kingdoms_sim/*.cpp)

add_library(kingdoms_sim STATIC ${KINGDOMS_SIM_SOURCES})
target_include_directories(kingdoms_sim PUBLIC include)
# The archive gets linked into PufferLib's _C.so (a shared library), so all
# objects must be position-independent.
set_target_properties(kingdoms_sim PROPERTIES POSITION_INDEPENDENT_CODE ON)
target_link_libraries(kingdoms_sim PUBLIC nlohmann_json::nlohmann_json)

# Path to the live game's JSON asset files (server/engine/assets), so item/npc/
# projectile/bomb balance never drifts out of sync with a hand-copied version.
get_filename_component(KINGDOMS_ROOT "${CMAKE_CURRENT_SOURCE_DIR}/.." ABSOLUTE)
target_compile_definitions(kingdoms_sim PUBLIC
  KINGDOMS_ASSETS_DIR="${KINGDOMS_ROOT}/server/engine/assets")

add_executable(rollout_demo apps/rollout_demo.cpp)
target_link_libraries(rollout_demo PRIVATE kingdoms_sim)
