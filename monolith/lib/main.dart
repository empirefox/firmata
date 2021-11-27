import 'package:flutter/material.dart';
import 'package:flutter_easyloading/flutter_easyloading.dart';
import 'package:get/get.dart';
import 'package:monolith/app/share/ext.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'app/data/services/storage.service.dart';
import 'app/data/services/transport.service.dart';
import 'app/routes/app_pages.dart';

void main() async {
  // This is required so ObjectBox can get the application directory
  // to store the database in.
  WidgetsFlutterBinding.ensureInitialized();

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
