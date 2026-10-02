// Copyright 2026 The cisort Authors.

#include "pch.h"
#include "widget.h"
#include "core/model.h"
#include <QtWidgets/QWidget>
#include <boost/asio.hpp>
#include <iostream>
#include <stdio.h>
#include <unistd.h>
#include <vector>

#include "llvm/ADT/StringRef.h"
#include <QString>
#include <algorithm>
#include <cmath>
#include <gtest/gtest.h>
#include <map>
// Needed for std::mutex on older toolchains.
#include <mutex>

#ifdef _WIN32
#include <windows.h>
#include <winsock2.h> // cisort: keep
#endif

namespace ui {
Widget::Widget() = default;
}  // namespace ui
