// Copyright 2026 The cisort Authors.

#include <vector>
#include "pch.h"
#include <QtWidgets/QWidget>
#include "widget.h"
#include <stdio.h>
#include <iostream>
#include "core/model.h"
#include <unistd.h>
#include <boost/asio.hpp>

#include <QString>
#include "llvm/ADT/StringRef.h"
#include <algorithm>
#include <map>
// Needed for std::mutex on older toolchains.
#include <mutex>
#include <cmath>
#include <gtest/gtest.h>

#ifdef _WIN32
#include <windows.h>
#include <winsock2.h> // cisort: keep
#endif

namespace ui {
Widget::Widget() = default;
}  // namespace ui
