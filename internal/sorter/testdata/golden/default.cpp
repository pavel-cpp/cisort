// Copyright 2026 The cisort Authors.

#include "pch.h"
#include "widget.h"
#include <stdio.h>
#include <unistd.h>
#include <iostream>
#include <vector>
#include <QtWidgets/QWidget>
#include <boost/asio.hpp>
#include "core/model.h"

#include <algorithm>
#include <cmath>
#include <map>
// Needed for std::mutex on older toolchains.
#include <mutex>
#include <QString>
#include <gtest/gtest.h>
#include "llvm/ADT/StringRef.h"

#ifdef _WIN32
#include <windows.h>
#include <winsock2.h> // cisort: keep
#endif

namespace ui {
Widget::Widget() = default;
}  // namespace ui
