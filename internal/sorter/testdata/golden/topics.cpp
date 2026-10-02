// Copyright 2026 The cisort Authors.

// Precompiled header
#include "pch.h"

// Main header
#include "widget.h"

// Containers
#include <map>
#include <vector>

// Streams and I/O
#include <iostream>

// Concurrency
// Needed for std::mutex on older toolchains.
#include <mutex>

// Algorithms and numerics
#include <algorithm>

// C library
#include <cmath>
#include <stdio.h>

// POSIX
#include <unistd.h>

// Third-party
#include <QString>
#include <QtWidgets/QWidget>
#include <boost/asio.hpp>
#include <gtest/gtest.h>

// Project
#include "core/model.h"
#include "llvm/ADT/StringRef.h"

#ifdef _WIN32
#include <windows.h>
#include <winsock2.h> // cisort: keep
#endif

namespace ui {
Widget::Widget() = default;
}  // namespace ui
