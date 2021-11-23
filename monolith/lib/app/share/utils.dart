import 'package:dartx/dartx.dart' show StringToIntOrNullExtension;
import 'package:get/get.dart' show Get, GetNavigation;

abstract class ShareUtils {
  static T? argOrParam<T extends ArgToParamer, S>(
    T? arg,
    S? param,
    ParamToArg<T, S> parse,
  ) {
    if (arg == null && param == null) return null;
    if (arg != null && param != null && arg.argToParam() != param.toString()) {
      return null;
    }
    arg ??= parse(param!);
    return arg;
  }

  static T? getxArgOrParam<T extends ArgToParamer>(
    String paramKey,
    ParamToArg<T, String> parse,
  ) {
    return argOrParam(Get.arguments as T?, Get.parameters[paramKey], parse);
  }

  static T? getxArgOrIdParam<T extends ArgToParamer>(
    ParamToArg<T, int> parse,
  ) {
    final id = Get.parameters['id']?.toIntOrNull();
    return argOrParam(Get.arguments as T?, id, parse);
  }
}

typedef ParamToArg<T, S> = T? Function(S param);

abstract class ArgToParamer {
  String argToParam();
}
