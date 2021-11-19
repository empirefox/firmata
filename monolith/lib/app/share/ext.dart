import 'dart:typed_data';

import 'package:dartx/dartx.dart';

extension ShareStringExt on String {
  String? get zeroAsNull => isEmpty ? null : this;
  String upperWords() => split('_').map((e) => e.capitalize()).join(' ');
}

extension ShareUint8ListExt on Uint8List {
  Uint8List? get zeroAsNull => isEmpty ? null : this;
}

extension ShareNumExt<T extends num> on T {
  T? get zeroAsNull => toDouble() == 0.0 ? null : this;
}
