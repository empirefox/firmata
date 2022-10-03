import 'dart:async';
import 'dart:developer';

import 'package:flutter/widgets.dart';
import 'package:get/get.dart' show Get, GetNavigation, GetxController, Inst;
import 'package:group_list_view/group_list_view.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/providers/transport.dart';
import 'package:monolith/app/data/services/storage.service.dart';
import 'package:monolith/app/data/services/transport.service.dart';
import 'package:monolith/app/routes/app_pages.dart';
import 'package:monolith/app/share/types.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';

import '../types/types.dart';

class PlanetController extends GetxController with WidgetsBindingObserver {
  final _ss = Get.find<StorageService>();
  final _ts = Get.find<TransportService>();

  Future<Transport>? _future;

  // requires future
  late final PlanetConfig config;
  // requires future
  late final Transport transport;
  List<Instance?> get instances => transport.instances;

  bool _appShown = true;
  StreamSubscription? _sub;

  @override
  void onInit() {
    super.onInit();
    WidgetsBinding.instance!.addObserver(this);
  }

  @override
  void onReady() {
    super.onReady();
  }

  @override
  void onClose() {
    _cancelListen();
    WidgetsBinding.instance!.removeObserver(this);
    super.onClose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    print('state changed');
    switch (state) {
      case AppLifecycleState.resumed:
        _appShown = true;
        break;
      case AppLifecycleState.inactive:
        _appShown = false;
        break;
      case AppLifecycleState.paused:
        _appShown = false;
        break;
      case AppLifecycleState.detached:
        _appShown = false;
        break;
    }
    _checkListenOrCancel();
  }

  // future must not be null to proccess.
  Future<Transport>? get future {
    if (_future == null) {
      PlanetConfig? config = _ss.getxPlanet();
      if (config == null) {
        return null;
      }

      this.config = config;

      _future = _ts.getOrCreate(config).then((t) {
        transport = t;
        _checkListenOrCancel();
        return t;
      });
    }
    return _future;
  }

  void _checkListenOrCancel() {
    if (_appShown) {
      _listen();
    } else {
      _cancelListen();
    }
  }

  void _listen() {
    if (_sub == null && _future != null) {
      _sub = transport.onServerMessage.listen((_) => update());
      update();
    }
  }

  void _cancelListen() {
    _sub?.cancel();
    _sub = null;
  }

  //
  // View
  //

  void notFound() => Get.offAllNamed(Routes.PLANETS);

  void about() => Get.toNamed(Routes.ABOUT);

  bool isInstanceAlive(VisibleGroupPin vgp) =>
      instances[vgp.pin.firmataIndex] != null;

  VisibleGroupPin groupPin(IndexPath index) =>
      transport.visibleGroupPins[index.section][index.index];

  //
  // GroupListView
  //

  int countOfItemInSection(int section) =>
      transport.visibleGroupPins[section].length;

  //
  // GetBuilderFilters
  //

  GetBuilderFilter<PlanetController> detectOrAliveFilter(
          VisibleGroupPin vgp, Instance_Pin pin, Instance_Pin? detect) =>
      detect != null ? detectFilter(vgp, pin, detect) : aliveFilter(vgp);

  GetBuilderFilter<PlanetController> aliveFilter(VisibleGroupPin vgp) =>
      (_) => instances[vgp.pin.firmataIndex]?.firmataIndex ?? -1;

  GetBuilderFilter<PlanetController> valueFilter(
          VisibleGroupPin vgp, Instance_Pin pin) =>
      (_) => instances[vgp.pin.firmataIndex] == null ? double.nan : pin.value;

  GetBuilderFilter<PlanetController> detectFilter(
          VisibleGroupPin vgp, Instance_Pin pin, Instance_Pin? detect) =>
      (_) => instances[vgp.pin.firmataIndex] == null
          ? double.nan
          : (detect?.value ?? pin.value);

  TriggerCallback onTriggerButton(VisibleGroupPin vgp) {
    return (int ms) => transport.triggerDigitalPin(
          group: vgp.group,
          gpin: vgp.gpin,
          realtimeTriggerMs: ms,
        );
  }

  //
  // actions
  //

  TriggerCallback onTriggerSwitch(VisibleGroupPin vgp, Instance_Pin pin) {
    return (int ms) {
      if (ms == 0) {
        return transport.setPinValue(
          pin: pin,
          group: vgp.group,
          gpin: vgp.gpin,
          value: vgp.getPin(transport)!.value ^ 1,
        );
      }
      return transport.triggerDigitalPin(
        group: vgp.group,
        gpin: vgp.gpin,
        realtimeTriggerMs: ms,
      );
    };
  }

  TriggerCallback onTriggerNumberWriter(VisibleGroupPin vgp, Instance_Pin pin) {
    return (int value) => transport.setPinValue(
          pin: pin,
          group: vgp.group,
          gpin: vgp.gpin,
          value: value,
        );
  }
}
