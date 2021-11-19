import 'package:flutter/material.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';

class PinError extends StatelessWidget {
  final Group_Pin groupPin;
  const PinError({Key? key, required this.groupPin}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Text('${groupPin.nick}(error)'),
      subtitle: Text(groupPin.desc),
      trailing: const Icon(Icons.block, color: Colors.yellow),
      enabled: false,
    );
  }
}
