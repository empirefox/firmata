import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart' show ExtensionSnackbar, Get;
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';

import '../types/types.dart';

// TODO add support: min, max, step.
class PinNumberWriter extends StatefulWidget {
  final Group_Pin groupPin;
  final Instance_Pin pin;
  final ValueCallback onTrigger;
  final bool enabled;
  const PinNumberWriter(
      {Key? key,
      required this.groupPin,
      required this.pin,
      required this.onTrigger,
      required this.enabled})
      : super(key: key);

  @override
  _PinNumberWriterState createState() => _PinNumberWriterState();
}

class _PinNumberWriterState extends State<PinNumberWriter> {
  Group_Pin get groupPin => widget.groupPin;
  Instance_Pin get pin => widget.pin;

  int get initialValue {
    final type = groupPin.numberWriter;
    return type.hasRecommend() ? type.recommend : pin.value;
  }

  TextEditingController? controller;

  @override
  void initState() {
    super.initState();
    controller ??= TextEditingController(text: initialValue.toString());
  }

  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Text('${groupPin.nick}: ${pin.value}'),
      subtitle: Text(groupPin.desc),
      trailing: TextField(
        controller: controller,
        decoration: InputDecoration(
          suffix: IconButton(
            onPressed: widget.enabled && controller!.text.trim().isNotEmpty
                ? () {
                    final value = int.tryParse(controller!.text);
                    if (value == null) {
                      Get.snackbar('${groupPin.nick} error',
                          'value must be integer, but got "${controller!.text}"');
                      return;
                    }
                    widget.onTrigger.call(value).then(
                          (_) => setState(() {}),
                          onError: (err) =>
                              Get.snackbar('${groupPin.nick} error', '$err'),
                        );
                  }
                : null,
            icon: const Icon(Icons.send),
          ),
        ),
        inputFormatters: [
          LengthLimitingTextInputFormatter(8,
              maxLengthEnforcement: MaxLengthEnforcement.enforced),
          FilteringTextInputFormatter.digitsOnly,
        ],
      ),
      enabled: widget.enabled,
    );
  }
}
