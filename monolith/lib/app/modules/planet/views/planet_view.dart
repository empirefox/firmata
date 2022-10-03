import 'dart:developer';

import 'package:flutter/material.dart';
import 'package:flutter_easyloading/flutter_easyloading.dart' show EasyLoading;
import 'package:get/get.dart' show Get, GetBuilder, Inst;
import 'package:group_list_view/group_list_view.dart';
import 'package:monolith/app/data/providers/transport.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';

import '../controllers/planet_controller.dart';
import 'pin_button.dart';
import 'pin_digitalreader.dart';
import 'pin_error.dart';
import 'pin_numberreader.dart';
import 'pin_numberwriter.dart';
import 'pin_switch.dart';

class PlanetView extends StatelessWidget {
  final controller = Get.find<PlanetController>();
  final nil = Container();

  PlanetView({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    if (controller.future == null) {
      return Center(
        child: Column(
          children: [
            Text('Planet not found'),
            _notFoundButton(),
          ],
        ),
      );
    }
    return FutureBuilder(
      future: controller.future,
      builder: (BuildContext context, AsyncSnapshot<Transport> snapshot) {
        final title = controller.config.viewName(snapshot.data?.config.nick);
        return Scaffold(
          appBar: AppBar(
            title: Text(title),
            actions: [
              IconButton(
                onPressed: controller.about,
                icon: const Icon(Icons.favorite),
              ),
            ],
          ),
          body: _futureBuilder(context, snapshot),
        );
      },
    );
  }

  Widget _futureBuilder(
      BuildContext context, AsyncSnapshot<Transport> snapshot) {
    switch (snapshot.connectionState) {
      case ConnectionState.none:
        return nil;
      case ConnectionState.waiting:
        EasyLoading.show(status: 'connecting...');
        return nil;
      case ConnectionState.active:
        return nil;
      case ConnectionState.done:
        EasyLoading.dismiss();
        if (snapshot.hasError) {
          return Center(
            child: Column(
              children: [
                Text(
                    'Failed to connect to ${controller.config.address}: ${snapshot.error}'),
                _notFoundButton(),
              ],
            ),
          );
        }
        return _listView(context);
    }
  }

  Widget _notFoundButton() {
    return MaterialButton(
      shape: CircleBorder(),
      color: Colors.blue,
      padding: EdgeInsets.all(20),
      onPressed: controller.notFound,
      child: Icon(Icons.replay),
    );
  }

  Widget _listView(BuildContext context) {
    final transport = controller.transport;
    final groups = transport.config.groups;
    return GroupListView(
      sectionsCount: groups.length,
      countOfItemInSection: controller.countOfItemInSection,
      separatorBuilder: (context, index) => const SizedBox(height: 5),
      sectionSeparatorBuilder: (context, section) => const SizedBox(height: 10),
      itemBuilder: _itemBuilder,
      groupHeaderBuilder: _groupHeaderBuilder,
    );
  }

  Widget _groupHeaderBuilder(BuildContext context, int section) {
    final transport = controller.transport;
    final groups = transport.config.groups;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 8),
      child: ListTile(
        title: Text(
          groups[section].name,
          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
        ),
        subtitle:
            groups[section].desc.isEmpty ? null : Text(groups[section].desc),
      ),
    );
  }

  Widget _itemBuilder(BuildContext context, IndexPath index) {
    final transport = controller.transport;
    final vgp = controller.groupPin(index);
    final pin = vgp.getPin(transport);
    if (pin == null) {
      return PinError(groupPin: vgp.pin);
    }
    switch (vgp.pin.whichType()) {
      case Group_Pin_Type.button:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(vgp),
          builder: (_) {
            return PinButton(
              groupPin: vgp.pin,
              onTrigger: controller.onTriggerButton(vgp),
              enabled: controller.isInstanceAlive(vgp),
            );
          },
        );
      case Group_Pin_Type.switch_21:
        final detect = vgp.getDetectPin(transport);
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.detectOrAliveFilter(vgp, pin, detect),
          builder: (_) {
            return PinSwitch(
              transport: transport,
              groupPin: vgp.pin,
              pin: pin,
              detect: detect,
              onTrigger: controller.onTriggerSwitch(vgp, pin),
              enabled: controller.isInstanceAlive(vgp),
            );
          },
        );
      case Group_Pin_Type.numberWriter:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(vgp),
          builder: (_) {
            return PinNumberWriter(
              groupPin: vgp.pin,
              pin: pin,
              onTrigger: controller.onTriggerNumberWriter(vgp, pin),
              enabled: controller.isInstanceAlive(vgp),
            );
          },
        );
      case Group_Pin_Type.digitalReader:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(vgp),
          builder: (_) {
            return PinDigitalReader(
              groupPin: vgp.pin,
              pin: pin,
              enabled: controller.isInstanceAlive(vgp),
            );
          },
        );
      case Group_Pin_Type.numberReader:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(vgp),
          builder: (_) {
            return PinNumberReader(
              groupPin: vgp.pin,
              pin: pin,
              enabled: controller.isInstanceAlive(vgp),
            );
          },
        );
      case Group_Pin_Type.hide:
        log('${vgp.pin.nick} error: hidden pins should not be handled');
        return PinError(groupPin: vgp.pin);
      case Group_Pin_Type.notSet:
        log('${vgp.pin.nick} bugs: type must be set');
        return PinError(groupPin: vgp.pin);
    }
  }
}
