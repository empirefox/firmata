import 'package:flutter/material.dart';
import 'package:flutter_easyloading/flutter_easyloading.dart';
import 'package:get/get.dart';
import 'package:monolith/app/share/ext.dart';
import 'package:package_info/package_info.dart';

import 'app/data/services/storage.service.dart';
import 'app/data/services/transport.service.dart';
import 'app/routes/app_pages.dart';

void main() async {
  final packageInfo =
      await Get.putAsync(PackageInfo.fromPlatform, permanent: true);
  await Get.putAsync(StorageService.create, permanent: true);
  Get.put(TransportService(), permanent: true);
  final _routeObserver = Get.put(RouteObserver<ModalRoute<void>>());

  runApp(
    GetMaterialApp(
      title: packageInfo.appName.upperWords(),
      initialRoute: AppPages.INITIAL,
      getPages: AppPages.routes,
      builder: EasyLoading.init(),
      navigatorObservers: [
        _routeObserver,
      ],
    ),
  );
}
