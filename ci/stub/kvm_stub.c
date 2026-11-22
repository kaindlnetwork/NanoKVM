/*
 * Host-side stub of libkvm.so for CI.
 *
 * The real libkvm.so is a riscv64 binary that drives the SG2002 video
 * pipeline. Unit tests run on x86_64/arm64 CI hosts, where the linker
 * cannot use it. This stub exports the same symbols with inert behaviour
 * so that packages importing NanoKVM-Server/common can be linked and
 * tested. It is never shipped.
 */
#include "kvm_vision.h"
#include <stddef.h>

void kvmv_init(uint8_t debug_info) { (void)debug_info; }
void kvmv_deinit(void) {}
int kvmv_read_img(uint16_t w, uint16_t h, uint8_t t, uint16_t q, uint16_t f,
                  kvmv_frame_t *frame) {
    (void)w; (void)h; (void)t; (void)q; (void)f;
    if (frame) { frame->data = NULL; frame->size = 0; }
    return IMG_NOT_EXIST;
}
void kvmv_release_frame(kvmv_frame_t *frame) { (void)frame; }
void kvmv_stop_stream(uint8_t type) { (void)type; }
void set_h264_gop(uint8_t gop) { (void)gop; }
void set_frame_detact(uint8_t interval) { (void)interval; }
uint8_t kvmv_hdmi_control(uint8_t enable) { (void)enable; return 0; }
uint8_t kvmv_hdmi_signal_active(void) { return 0; }
