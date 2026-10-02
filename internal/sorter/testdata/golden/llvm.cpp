// Copyright 2026 The cisort Authors.

#include "pch.h"
#include "widget.h"
#include "core/model.h"
#include "llvm/ADT/StringRef.h"
#include <QString>
#include <QtWidgets/QWidget>
#include <algorithm>
#include <boost/asio.hpp>
#include <cmath>
#include <gtest/gtest.h>
#include <iostream>
#include <map>
// Needed for std::mutex on older toolchains.
#include <mutex>
#include <stdio.h>
#include <unistd.h>
#include <vector>

#ifdef _WIN32
#include <windows.h>
#include <winsock2.h> // cisort: keep
#endif

namespace ui {
Widget::Widget() = default;
}  // namespace ui
